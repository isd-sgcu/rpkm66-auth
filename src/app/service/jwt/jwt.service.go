package jwt

import (
	"time"

	_jwt "github.com/golang-jwt/jwt/v4"
	dto "github.com/isd-sgcu/rpkm66-auth/src/app/dto/auth"
	entity "github.com/isd-sgcu/rpkm66-auth/src/app/entity/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/src/pkg/strategy"
	"github.com/pkg/errors"
)

type serviceImpl struct {
	conf     cfgldr.Jwt
	strategy strategy.JwtStrategy
}

func NewJwtService(conf cfgldr.Jwt, strategy strategy.JwtStrategy) *serviceImpl {
	return &serviceImpl{
		conf:     conf,
		strategy: strategy,
	}
}

func (s *serviceImpl) SignAuth(in *entity.Auth) (string, error) {
	payloads := &dto.TokenPayloadAuth{
		RegisteredClaims: _jwt.RegisteredClaims{
			Issuer:    s.conf.Issuer,
			ExpiresAt: _jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(s.conf.ExpiresIn))),
			IssuedAt:  _jwt.NewNumericDate(time.Now()),
		},
		UserId: in.UserID,
	}
	token := _jwt.NewWithClaims(_jwt.SigningMethodHS256, payloads)

	tokenStr, err := token.SignedString([]byte(s.conf.Secret))
	if err != nil {
		return "", errors.New("Error while signing the token")
	}

	return tokenStr, nil
}

func (s *serviceImpl) VerifyAuth(token string) (*_jwt.Token, error) {
	return _jwt.Parse(token, s.strategy.AuthDecode)
}

func (s *serviceImpl) GetConfig() *cfgldr.Jwt {
	return &s.conf
}
