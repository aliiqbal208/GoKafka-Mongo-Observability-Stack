package usecase

import (
	"context"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/user"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// userUseCase implements user.UseCase
type userUseCase struct {
	userRepo user.Repository
	log      logger.Logger
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo user.Repository, log logger.Logger) user.UseCase {
	return &userUseCase{
		userRepo: userRepo,
		log:      log,
	}
}

// Signup creates a new user account
func (uc *userUseCase) Signup(ctx context.Context, req *models.SignupRequest) (*models.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.Signup")
	defer span.Finish()

	// Check if email already exists
	exists, err := uc.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, errors.Wrap(err, "userUseCase.Signup.EmailExists")
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Create new user
	newUser := models.NewUser(req)

	// Hash password
	if err := newUser.HashPassword(); err != nil {
		return nil, errors.Wrap(err, "userUseCase.Signup.HashPassword")
	}

	// Save to database
	createdUser, err := uc.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, errors.Wrap(err, "userUseCase.Signup.Create")
	}

	uc.log.Infof("User created: %s", createdUser.Email)
	return createdUser.ToResponse(), nil
}

// Login validates user credentials
func (uc *userUseCase) Login(ctx context.Context, req *models.LoginRequest) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.Login")
	defer span.Finish()

	// Get user by email
	u, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Wrap(err, "userUseCase.Login.GetByEmail")
	}
	if u == nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare password
	if !u.ComparePassword(req.Password) {
		return nil, errors.New("invalid email or password")
	}

	uc.log.Infof("User logged in: %s", u.Email)
	return u, nil
}

// GetByID retrieves a user by ID
func (uc *userUseCase) GetByID(ctx context.Context, userID string) (*models.UserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userUseCase.GetByID")
	defer span.Finish()

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "userUseCase.GetByID")
	}

	return u.ToResponse(), nil
}
