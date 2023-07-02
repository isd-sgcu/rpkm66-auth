package token

import (
	dto "github.com/isd-sgcu/rpkm66-auth/internal/dto/auth"
	entity "github.com/isd-sgcu/rpkm66-auth/internal/entity/auth"
	proto "github.com/isd-sgcu/rpkm66-auth/internal/proto/rpkm66/auth/auth/v1"
	token_svc "github.com/isd-sgcu/rpkm66-auth/internal/service/token"
	cache_repo "github.com/isd-sgcu/rpkm66-auth/pkg/repository/cache"
	jwt_svc "github.com/isd-sgcu/rpkm66-auth/pkg/service/jwt"
)

type Service interface {
	CreateCredentials(auth *entity.Auth, secret string) (*proto.Credential, error)
	Validate(token string) (*dto.UserCredential, error)
	CreateRefreshToken() string
}

func NewTokenService(jwtService jwt_svc.Service, cacheRepository cache_repo.Repository) Service {
	return token_svc.NewService(jwtService, cacheRepository)
}
