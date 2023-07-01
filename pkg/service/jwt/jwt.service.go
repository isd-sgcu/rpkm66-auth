package jwt

import (
	_jwt "github.com/golang-jwt/jwt/v4"
	"github.com/isd-sgcu/rpkm66-auth/cfgldr"
	entity "github.com/isd-sgcu/rpkm66-auth/internal/entity/auth"
	"github.com/isd-sgcu/rpkm66-auth/internal/service/jwt"
	"github.com/isd-sgcu/rpkm66-auth/pkg/strategy"
)

type Service interface {
	SignAuth(in *entity.Auth) (string, error)
	VerifyAuth(token string) (*_jwt.Token, error)
	GetConfig() *cfgldr.Jwt
}

func NewJwtService(conf cfgldr.Jwt, strategy strategy.JwtStrategy) Service {
	return jwt.NewJwtService(conf, strategy)
}
