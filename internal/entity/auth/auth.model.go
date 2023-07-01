package auth

import (
	"github.com/isd-sgcu/rpkm66-auth/internal/entity"
)

type Auth struct {
	entity.Base
	UserID       string `json:"user_id" gorm:"index:,unique"`
	Role         string `json:"role" gorm:"type:text"`
	RefreshToken string `json:"refresh_token" gorm:"index"`
}
