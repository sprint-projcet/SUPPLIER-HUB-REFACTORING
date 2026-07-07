package repositories

import (
	"supplierhub-backend/models"

	"gorm.io/gorm"
)

// ProductRepository provides database access for Product entities
type ProductRepository interface {
	Insert(product *models.Product) error
}

type productRepo struct {
	db *gorm.DB
}

// NewProductRepository creates a new ProductRepository instance
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Insert(product *models.Product) error {
	return r.db.Create(product).Error
}
