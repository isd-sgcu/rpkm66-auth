package auth

import (
	entity "github.com/isd-sgcu/rpkm66-auth/internal/entity/auth"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByUserID(uid string, result *entity.Auth) error {
	return r.db.First(&result, "user_id = ?", uid).Error
}

func (r *Repository) FindByRefreshToken(refreshToken string, result *entity.Auth) error {
	return r.db.First(&result, "refresh_token = ?", refreshToken).Error
}

func (r *Repository) Create(auth *entity.Auth) error {
	return r.db.Create(&auth).Error
}

func (r *Repository) Update(id string, auth *entity.Auth) error {
	return r.db.Where(id, "id = ?", id).Updates(&auth).First(&auth, "id = ?", id).Error
}
