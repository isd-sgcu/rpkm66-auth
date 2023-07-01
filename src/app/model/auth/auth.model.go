package auth

import (
	"github.com/isd-sgcu/rpkm66-auth/src/app/model"
)

type Auth struct {
	model.Base
	UserID       string `json:"user_id" gorm:"index:,unique"`
	Role         string `json:"role" gorm:"type:text"`
	RefreshToken string `json:"refresh_token" gorm:"index"`
}
