package user

import (
	"context"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
)

// UseCase defines the user use case interface
type UseCase interface {
	Signup(ctx context.Context, req *models.SignupRequest) (*models.UserResponse, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.UserResponse, error)
}
