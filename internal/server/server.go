package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// Core dependencies
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// gRPC middleware
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	// Fiber
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"

	// Internal packages
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/config"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/interceptors"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/middlewares"

	// IMPORTANT: rename your local “internal/product/delivery/grpc” package to avoid name collision:
	// Cart module
	cartHttpV1 "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart/delivery/http/v1"
	cartRepository "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart/repository"
	cartUsecase "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart/usecase"

	// Order module
	orderHttpV1 "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order/delivery/http/v1"
	orderKafka "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order/delivery/kafka"
	orderRepository "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order/repository"
	orderUsecase "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order/usecase"

	// Product module
	productGrpc "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product/delivery/grpc"
	productsHttpV1 "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product/delivery/http/v1"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product/delivery/kafka"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product/repository"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product/usecase"

	// User module
	userHttpV1 "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/user/delivery/http/v1"
	userRepository "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/user/repository"
	userUsecase "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/user/usecase"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/jwt"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
	productsService "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/proto/product"
)

const (
	maxHeaderBytes  = 1 << 20
	gzipLevel       = 5
	stackSize       = 1 << 10
	csrfTokenHeader = "X-CSRF-Token"
	bodyLimit       = 2 * 1024 * 1024 // 2MB

	kafkaGroupID = "products_group"
)

type server struct {
	log     logger.Logger
	cfg     *config.Config
	tracer  opentracing.Tracer
	mongoDB *mongo.Client
	app     *fiber.App
	redis   *redis.Client
}

// NewServer constructs our main server object.
func NewServer(
	log logger.Logger,
	cfg *config.Config,
	tracer opentracing.Tracer,
	mongoDB *mongo.Client,
	redis *redis.Client,
) *server {
	return &server{
		log:     log,
		cfg:     cfg,
		tracer:  tracer,
		mongoDB: mongoDB,
		app:     fiber.New(fiber.Config{BodyLimit: bodyLimit}),
		redis:   redis,
	}
}

// Run starts up everything (gRPC, Kafka consumers, and the HTTP server).
func (s *server) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Validator for incoming requests.
	validate := validator.New()

	// Kafka Producer.
	productsProducer := kafka.NewProductsProducer(s.log, s.cfg)
	productsProducer.Run()
	defer productsProducer.Close()

	// Orders Kafka Producer.
	ordersProducer := orderKafka.NewOrdersProducer(s.log, s.cfg)
	ordersProducer.Run()
	defer ordersProducer.Close()

	// Setup Product Repos & UseCase.
	productMongoRepo := repository.NewProductMongoRepo(s.mongoDB)
	productRedisRepo := repository.NewProductRedisRepository(s.redis)
	productUC := usecase.NewProductUC(productMongoRepo, productRedisRepo, s.log, productsProducer)

	// Setup Cart Repos & UseCase.
	db := s.mongoDB.Database(s.cfg.MongoDB.DB)
	cartMongoRepo := cartRepository.NewCartMongoRepository(s.log, db)
	cartRedisRepo := cartRepository.NewCartRedisRepository(s.log, s.redis)
	cartUC := cartUsecase.NewCartUseCase(s.log, cartMongoRepo, cartRedisRepo, productMongoRepo)

	// Setup Order Repos & UseCase.
	orderMongoRepo := orderRepository.NewOrderMongoRepository(s.log, db)
	orderRedisRepo := orderRepository.NewOrderRedisRepository(s.log, s.redis)
	orderUC := orderUsecase.NewOrderUseCase(s.log, orderMongoRepo, orderRedisRepo, cartUC, productMongoRepo)

	// Setup JWT Manager.
	jwtExpireHours := s.cfg.Server.JWTExpireHours
	if jwtExpireHours == 0 {
		jwtExpireHours = 24 // Default to 24 hours
	}
	jwtMgr := jwt.NewManager(s.cfg.Server.JWTSecret, jwtExpireHours)

	// Setup User Repos & UseCase.
	userMongoRepo := userRepository.NewUserMongoRepo(db)
	userUC := userUsecase.NewUserUseCase(userMongoRepo, s.log)

	// Interceptors & Middlewares.
	im := interceptors.NewInterceptorManager(s.log, s.cfg)
	mw := middlewares.NewMiddlewareManager(s.log, s.cfg, jwtMgr)

	// gRPC Server setup.
	grpcAddr := s.cfg.Server.Port
	if !strings.HasPrefix(grpcAddr, ":") {
		grpcAddr = ":" + grpcAddr
	}
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
			Timeout:           s.cfg.Server.Timeout * time.Second,
			MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
			Time:              s.cfg.Server.Timeout * time.Minute,
		}),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpcrecovery.UnaryServerInterceptor(),
			im.Logger,
		),
	)

	// Register our product service implementation with gRPC.
	prodSvc := productGrpc.NewProductService(s.log, productUC, validate)
	productsService.RegisterProductsServiceServer(grpcServer, prodSvc)
	grpc_prometheus.Register(grpcServer)

	// HTTP routes for /api/v1/products (this is separate from the top-level routes in http.go).
	v1 := s.app.Group("/api/v1")
	v1.Use(mw.Metrics)

	productHandlers := productsHttpV1.NewProductHandlers(s.log, productUC, validate, v1.Group("/products"), mw)
	productHandlers.MapRoutes()

	// Cart HTTP handlers.
	cartHandlers := cartHttpV1.NewCartHandlers(s.log, cartUC, validate, v1.Group("/cart"), mw)
	cartHandlers.MapRoutes()

	// Order HTTP handlers.
	orderHandlers := orderHttpV1.NewOrderHandlers(s.log, orderUC, v1.Group("/orders"))
	orderHandlers.MapRoutes()

	// User/Auth HTTP handlers.
	userHandlers := userHttpV1.NewUserHandlers(userUC, jwtMgr, s.log)
	userHttpV1.MapAuthRoutes(v1, userHandlers, mw)

	// Kafka Consumer Group for Products.
	productsCG := kafka.NewProductsConsumerGroup(
		s.cfg.Kafka.Brokers,
		kafkaGroupID,
		s.log,
		s.cfg,
		productUC,
		validate,
	)

	// Kafka Consumer Group for Orders.
	ordersCG := orderKafka.NewOrdersConsumerGroup(
		s.cfg.Kafka.Brokers,
		orderKafka.OrderConsumerGroup,
		s.log,
		s.cfg,
		orderUC,
		productMongoRepo,
		validate,
	)

	// Start gRPC in background.
	go func() {
		s.log.Infof("gRPC server listening on %s", grpcAddr)
		if err := grpcServer.Serve(listener); err != nil {
			s.log.Errorf("gRPC server error: %v", err)
			cancel()
		}
	}()

	// Start Kafka consumers in background.
	go productsCG.RunConsumers(ctx, cancel)
	go ordersCG.RunConsumers(ctx, cancel)

	// Start a separate metrics server in background (optional).
	go func() {
		metricsApp := fiber.New()
		metricsApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
		s.log.Infof("Metrics server on %s", s.cfg.Metrics.Port)
		addr := s.cfg.Metrics.Port
		if !strings.HasPrefix(addr, ":") {
			addr = ":" + addr
		}
		if err := metricsApp.Listen(addr); err != nil {
			s.log.Errorf("Metrics server: %v", err)
			cancel()
		}
	}()

	// Start the main HTTP server (defined in http.go), in background.
	go func() {
		if errHTTP := s.runHTTPServer(); errHTTP != nil {
			s.log.Errorf("HTTP server error: %v", errHTTP)
			cancel()
		}
	}()

	// Wait for OS signal or context cancellation.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		s.log.Warnf("Received signal: %v, shutting down...", sig)
	case <-ctx.Done():
		s.log.Warnf("Context canceled, shutting down servers...")
	}

	// Graceful shutdown attempts:
	grpcServer.GracefulStop()
	if err := s.app.Shutdown(); err != nil {
		s.log.Errorf("Error on shutting down HTTP server: %v", err)
	}
	_ = listener.Close()

	s.log.Info("Server exited properly")
	return nil
}
