package strategy

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/isd-sgcu/rpkm66-auth/internal/strategy"
)

type JwtStrategy interface {
	AuthDecode(token *jwt.Token) (interface{}, error)
}

func NewJwtStrategy(secret string) JwtStrategy {
	return strategy.NewJwtStrategy(secret)
}
