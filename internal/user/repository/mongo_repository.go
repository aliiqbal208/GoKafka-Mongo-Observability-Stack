package repository

import (
	"context"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/user"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const usersCollection = "users"

// userMongoRepo implements user.Repository
type userMongoRepo struct {
	mongoDB *mongo.Database
}

// NewUserMongoRepo creates a new user MongoDB repository
func NewUserMongoRepo(mongoDB *mongo.Database) user.Repository {
	return &userMongoRepo{mongoDB: mongoDB}
}

// Create creates a new user in MongoDB
func (r *userMongoRepo) Create(ctx context.Context, u *models.User) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userMongoRepo.Create")
	defer span.Finish()

	collection := r.mongoDB.Collection(usersCollection)

	result, err := collection.InsertOne(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "userMongoRepo.Create.InsertOne")
	}

	u.ID = result.InsertedID.(primitive.ObjectID)
	return u, nil
}

// GetByID retrieves a user by ID
func (r *userMongoRepo) GetByID(ctx context.Context, userID string) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userMongoRepo.GetByID")
	defer span.Finish()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.Wrap(err, "userMongoRepo.GetByID.ObjectIDFromHex")
	}

	collection := r.mongoDB.Collection(usersCollection)

	var u models.User
	if err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, errors.Wrap(err, "userMongoRepo.GetByID.FindOne")
	}

	return &u, nil
}

// GetByEmail retrieves a user by email
func (r *userMongoRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userMongoRepo.GetByEmail")
	defer span.Finish()

	collection := r.mongoDB.Collection(usersCollection)

	var u models.User
	if err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil to indicate user not found (for login checks)
		}
		return nil, errors.Wrap(err, "userMongoRepo.GetByEmail.FindOne")
	}

	return &u, nil
}

// EmailExists checks if an email already exists
func (r *userMongoRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "userMongoRepo.EmailExists")
	defer span.Finish()

	collection := r.mongoDB.Collection(usersCollection)

	count, err := collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, errors.Wrap(err, "userMongoRepo.EmailExists.CountDocuments")
	}

	return count > 0, nil
}
