package v1

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/middlewares"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product"
	httpErrors "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/http_errors"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/utils"
)

type productHandlers struct {
	log       logger.Logger
	productUC product.UseCase
	validate  *validator.Validate
	group     fiber.Router
	mw        middlewares.MiddlewareManager
}

// NewProductHandlers constructor
func NewProductHandlers(
	log logger.Logger,
	productUC product.UseCase,
	validate *validator.Validate,
	group fiber.Router,
	mw middlewares.MiddlewareManager,
) *productHandlers {
	return &productHandlers{log: log, productUC: productUC, validate: validate, group: group, mw: mw}
}

// CreateProduct Create product
// @Tags Products
// @Summary Create new product
// @Description Create new single product
// @Accept json
// @Produce json
// @Param product body models.Product true "Product data"
// @Success 201 {object} models.Product
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /products [post]
func (p *productHandlers) CreateProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "productHandlers.Create")
		defer span.Finish()
		createRequests.Inc()

		var prod models.Product
		if err := c.BodyParser(&prod); err != nil {
			p.log.Errorf("c.BodyParser: %v", err)
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := p.validate.StructCtx(ctx, &prod); err != nil {
			p.log.Errorf("validate.StructCtx: %v", err)
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := p.productUC.PublishCreate(ctx, &prod); err != nil {
			p.log.Errorf("productUC.PublishCreate: %v", err)
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.SendStatus(http.StatusCreated)
	}
}

// UpdateProduct Update product
// @Tags Products
// @Summary Update single product
// @Description Update single product by id
// @Accept json
// @Produce json
// @Param product_id path string true "product id"
// @Param product body models.Product true "Product data"
// @Success 200 {object} models.Product
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /products/{product_id} [put]
func (p *productHandlers) UpdateProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "productHandlers.Update")
		defer span.Finish()
		updateRequests.Inc()

		var prod models.Product
		if err := c.BodyParser(&prod); err != nil {
			p.log.Errorf("c.BodyParser: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		prodID, err := primitive.ObjectIDFromHex(c.Params("product_id"))
		if err != nil {
			p.log.Errorf("primitive.ObjectIDFromHex: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}
		prod.ProductID = prodID

		if err := p.validate.StructCtx(ctx, &prod); err != nil {
			p.log.Errorf("validate.StructCtx: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := p.productUC.PublishUpdate(ctx, &prod); err != nil {
			p.log.Errorf("productUC.PublishUpdate: %v", err)
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.SendStatus(http.StatusOK)
	}
}

// GetByIDProduct Get product by id
// @Tags Products
// @Summary Get product by id
// @Description Get single product by id
// @Accept json
// @Produce json
// @Param product_id path string true "product id"
// @Success 200 {object} models.Product
// @Router /products/{product_id} [get]
func (p *productHandlers) GetByIDProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "productHandlers.GetByID")
		defer span.Finish()
		getByIdRequests.Inc()

		objectID, err := primitive.ObjectIDFromHex(c.Params("product_id"))
		if err != nil {
			p.log.Errorf("primitive.ObjectIDFromHex: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		prod, err := p.productUC.GetByID(ctx, objectID)
		if err != nil {
			p.log.Errorf("productUC.GetByID: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(prod)
	}
}

// GetAllProducts Get all products
// @Tags Products
// @Summary Get all products
// @Description Get all products with pagination
// @Accept json
// @Produce json
// @Param page query string false "page number" default(1)
// @Param size query string false "number of elements" default(10)
// @Success 200 {object} models.ProductsList
// @Router /products [get]
func (p *productHandlers) GetAllProducts() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "productHandlers.GetAll")
		defer span.Finish()
		getByIdRequests.Inc()

		// Parse query parameters with defaults
		page := 1
		size := 10

		if pageParam := c.Query("page"); pageParam != "" {
			if parsedPage, err := strconv.Atoi(pageParam); err == nil && parsedPage > 0 {
				page = parsedPage
			}
		}

		if sizeParam := c.Query("size"); sizeParam != "" {
			if parsedSize, err := strconv.Atoi(sizeParam); err == nil && parsedSize > 0 {
				size = parsedSize
			}
		}

		pq := utils.NewPaginationQuery(size, page)
		result, err := p.productUC.GetAll(ctx, pq)
		if err != nil {
			p.log.Errorf("productUC.GetAll: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(result)
	}
}

// SearchProduct Search product
// @Tags Products
// @Summary Search product
// @Description Search product by name or description
// @Accept json
// @Produce json
// @Param search query string false "search text"
// @Param page query string false "page number"
// @Param size query string false "number of elements"
// @Success 200 {object} models.ProductsList
// @Router /products/search [get]
func (p *productHandlers) SearchProduct() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "productHandlers.Search")
		defer span.Finish()
		searchRequests.Inc()

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			p.log.Errorf("strconv.Atoi: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}
		size, err := strconv.Atoi(c.Query("size"))
		if err != nil {
			p.log.Errorf("strconv.Atoi: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		pq := utils.NewPaginationQuery(size, page)
		result, err := p.productUC.Search(ctx, c.Query("search"), pq)
		if err != nil {
			p.log.Errorf("productUC.Search: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(result)
	}
}
