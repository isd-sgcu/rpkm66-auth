package jwt

import (
	_jwt "github.com/golang-jwt/jwt/v4"
	entity "github.com/isd-sgcu/rpkm66-auth/src/app/entity/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/app/service/jwt"
	"github.com/isd-sgcu/rpkm66-auth/src/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/src/pkg/strategy"
)

type Service interface {
	SignAuth(in *entity.Auth) (string, error)
	VerifyAuth(token string) (*_jwt.Token, error)
	GetConfig() *cfgldr.Jwt
}

func NewJwtService(conf cfgldr.Jwt, strategy strategy.JwtStrategy) Service {
	return jwt.NewJwtService(conf, strategy)
}
