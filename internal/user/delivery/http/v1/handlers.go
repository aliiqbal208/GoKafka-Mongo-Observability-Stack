package v1

import (
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/middlewares"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/user"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/jwt"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/opentracing/opentracing-go"
)

// userHandlers implements user.Handlers
type userHandlers struct {
	userUC     user.UseCase
	jwtManager *jwt.Manager
	log        logger.Logger
	validate   *validator.Validate
}

// NewUserHandlers creates new user handlers
func NewUserHandlers(userUC user.UseCase, jwtManager *jwt.Manager, log logger.Logger) user.Handlers {
	return &userHandlers{
		userUC:     userUC,
		jwtManager: jwtManager,
		log:        log,
		validate:   validator.New(),
	}
}

// Signup godoc
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.SignupRequest true "Signup request"
// @Success      201 {object} models.UserResponse
// @Failure      400 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/v1/auth/signup [post]
func (h *userHandlers) Signup() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.UserContext(), "userHandlers.Signup")
		defer span.Finish()

		var req models.SignupRequest
		if err := c.BodyParser(&req); err != nil {
			h.log.Errorf("Signup.BodyParser: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if err := h.validate.Struct(req); err != nil {
			h.log.Errorf("Signup.Validate: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		userResp, err := h.userUC.Signup(ctx, &req)
		if err != nil {
			h.log.Errorf("Signup.UseCase: %v", err)
			if err.Error() == "email already registered" {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(userResp)
	}
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user and return JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Login request"
// @Success      200 {object} models.LoginResponse
// @Failure      400 {object} map[string]string
// @Failure      401 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /api/v1/auth/login [post]
func (h *userHandlers) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.UserContext(), "userHandlers.Login")
		defer span.Finish()

		var req models.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			h.log.Errorf("Login.BodyParser: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if err := h.validate.Struct(req); err != nil {
			h.log.Errorf("Login.Validate: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		u, err := h.userUC.Login(ctx, &req)
		if err != nil {
			h.log.Errorf("Login.UseCase: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}

		// Generate JWT token
		token, err := h.jwtManager.GenerateToken(u.ID.Hex(), u.Email, u.Name)
		if err != nil {
			h.log.Errorf("Login.GenerateToken: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		return c.Status(fiber.StatusOK).JSON(&models.LoginResponse{
			User:  u.ToResponse(),
			Token: token,
		})
	}
}

// Logout godoc
// @Summary      Logout user
// @Description  Logout user (client should discard token)
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /api/v1/auth/logout [post]
func (h *userHandlers) Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Logout successful. Please discard your token.",
		})
	}
}

// GetCurrentUser godoc
// @Summary      Get current user
// @Description  Get the currently authenticated user
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.UserResponse
// @Failure      401 {object} map[string]string
// @Router       /api/v1/auth/me [get]
func (h *userHandlers) GetCurrentUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.UserContext(), "userHandlers.GetCurrentUser")
		defer span.Finish()

		userID, ok := c.Locals(middlewares.UserIDKey).(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Not authenticated",
			})
		}

		userResp, err := h.userUC.GetByID(ctx, userID)
		if err != nil {
			h.log.Errorf("GetCurrentUser.GetByID: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get user",
			})
		}

		return c.Status(fiber.StatusOK).JSON(userResp)
	}
}
