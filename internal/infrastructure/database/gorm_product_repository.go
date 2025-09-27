package database

import (
	"ecs/internal/domain/product"

	"gorm.io/gorm"
)

type GormProductRepository struct {
	db *gorm.DB
}

func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

func (r *GormProductRepository) Save(p *product.Product) error {
	return r.db.Create(p).Error
}

func (r *GormProductRepository) FindAll() ([]product.Product, error) {
	var products []product.Product
	err := r.db.Find(&products).Error
	return products, err
}
