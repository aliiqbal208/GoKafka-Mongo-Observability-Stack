package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"

	productsService "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/proto/product"
)

// Product models
// @Description Product information
type Product struct {
	ProductID   primitive.ObjectID `json:"productId" bson:"_id,omitempty" swaggerignore:"true"`
	CategoryID  primitive.ObjectID `json:"categoryId,omitempty" bson:"categoryId,omitempty" swaggerignore:"true"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty" validate:"required,min=3,max=250" example:"iPhone 15 Pro"`
	Description string             `json:"description,omitempty" bson:"description,omitempty" validate:"required,min=3,max=500" example:"Latest Apple smartphone with A17 Pro chip"`
	Price       float64            `json:"price,omitempty" bson:"price,omitempty" validate:"required" example:"999.99"`
	ImageURL    *string            `json:"imageUrl,omitempty" bson:"imageUrl,omitempty" example:"https://example.com/image.jpg"`
	Photos      []string           `json:"photos,omitempty" bson:"photos,omitempty"`
	Quantity    int64              `json:"quantity,omitempty" bson:"quantity,omitempty" validate:"required" example:"100"`
	Rating      int                `json:"rating,omitempty" bson:"rating,omitempty" validate:"required,min=0,max=10" example:"8"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt,omitempty" swaggerignore:"true"`
}

func (p *Product) GetImage() string {
	var img string
	if p.ImageURL != nil {
		img = *p.ImageURL
	}
	return img
}

// ToProto Convert product to proto
func (p *Product) ToProto() *productsService.Product {
	return &productsService.Product{
		ProductID:   p.ProductID.String(),
		CategoryID:  p.CategoryID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		ImageURL:    p.GetImage(),
		Photos:      p.Photos,
		Quantity:    p.Quantity,
		Rating:      int64(p.Rating),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

// ProductFromProto Get Product from proto
func ProductFromProto(product *productsService.Product) (*Product, error) {
	prodID, err := primitive.ObjectIDFromHex(product.GetCategoryID())
	if err != nil {
		return nil, err
	}
	catID, err := primitive.ObjectIDFromHex(product.GetCategoryID())
	if err != nil {
		return nil, err
	}

	return &Product{
		ProductID:   prodID,
		CategoryID:  catID,
		Name:        product.GetName(),
		Description: product.GetDescription(),
		Price:       product.GetPrice(),
		ImageURL:    &product.ImageURL,
		Photos:      product.GetPhotos(),
		Quantity:    product.GetQuantity(),
		Rating:      int(product.GetRating()),
		CreatedAt:   product.GetCreatedAt().AsTime(),
		UpdatedAt:   product.GetUpdatedAt().AsTime(),
	}, nil
}

// ProductsList All Products response with pagination
type ProductsList struct {
	TotalCount int64      `json:"totalCount"`
	TotalPages int64      `json:"totalPages"`
	Page       int64      `json:"page"`
	Size       int64      `json:"size"`
	HasMore    bool       `json:"hasMore"`
	Products   []*Product `json:"products"`
}

// ToProtoList convert products list to proto
func (p *ProductsList) ToProtoList() []*productsService.Product {
	productsList := make([]*productsService.Product, 0, len(p.Products))
	for _, product := range p.Products {
		productsList = append(productsList, product.ToProto())
	}
	return productsList
}
