package auth

import (
	"github.com/isd-sgcu/rnkm65-auth/src/app/model"
)

type Auth struct {
	model.Base
	UserID       string `json:"user_id" gorm:"index:,unique"`
	Role         string `json:"role" gorm:"type:tinytext"`
	RefreshToken string `json:"refresh_token" gorm:"index"`
}
