package auth

import (
	entity "github.com/isd-sgcu/rpkm66-auth/internal/entity/auth"
	auth_repo "github.com/isd-sgcu/rpkm66-auth/internal/repository/auth"
	"gorm.io/gorm"
)

type Repository interface {
	FindByUserID(uid string, result *entity.Auth) error
	FindByRefreshToken(refreshToken string, result *entity.Auth) error
	Create(auth *entity.Auth) error
	Update(id string, auth *entity.Auth) error
}

func NewRepository(db *gorm.DB) Repository {
	return auth_repo.NewRepository(db)
}
