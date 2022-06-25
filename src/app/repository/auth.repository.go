package repository

import (
	model "github.com/isd-sgcu/rnkm65-auth/src/app/model/auth"
	"gorm.io/gorm"
)

type Repository struct {
	db gorm.DB
}

func NewRepository(db gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByRefreshToken(refreshToken string, result *model.Auth) error {
	return r.db.First(&result, "refresh_token = ?", refreshToken).Error
}

func (r *Repository) Create(auth *model.Auth) error {
	return r.db.Create(&auth).Error
}

func (r *Repository) Update(id string, auth *model.Auth) error {
	return r.db.Where(id).Updates(&auth).First(&auth).Error
}
