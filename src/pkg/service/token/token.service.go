package token

import (
	dto "github.com/isd-sgcu/rpkm66-auth/src/app/dto/auth"
	entity "github.com/isd-sgcu/rpkm66-auth/src/app/entity/auth"
	token_svc "github.com/isd-sgcu/rpkm66-auth/src/app/service/token"
	cache_repo "github.com/isd-sgcu/rpkm66-auth/src/pkg/repository/cache"
	jwt_svc "github.com/isd-sgcu/rpkm66-auth/src/pkg/service/jwt"
	"github.com/isd-sgcu/rpkm66-auth/src/proto"
)

type Service interface {
	CreateCredentials(auth *entity.Auth, secret string) (*proto.Credential, error)
	Validate(token string) (*dto.UserCredential, error)
	CreateRefreshToken() string
}

func NewTokenService(jwtService jwt_svc.Service, cacheRepository cache_repo.Repository) Service {
	return token_svc.NewService(jwtService, cacheRepository)
}
