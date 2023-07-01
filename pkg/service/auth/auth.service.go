package auth

import (
	"context"

	"github.com/isd-sgcu/rpkm66-auth/cfgldr"
	entity "github.com/isd-sgcu/rpkm66-auth/internal/entity/auth"
	"github.com/isd-sgcu/rpkm66-auth/internal/service/auth"
	"github.com/isd-sgcu/rpkm66-auth/pkg/client/chula_sso"
	auth_repo "github.com/isd-sgcu/rpkm66-auth/pkg/repository/auth"
	token_svc "github.com/isd-sgcu/rpkm66-auth/pkg/service/token"
	user_svc "github.com/isd-sgcu/rpkm66-auth/pkg/service/user"
	"github.com/isd-sgcu/rpkm66-auth/proto"
)

type Service interface {
	VerifyTicket(_ context.Context, req *proto.VerifyTicketRequest) (*proto.VerifyTicketResponse, error)
	Validate(_ context.Context, req *proto.ValidateRequest) (*proto.ValidateResponse, error)
	RefreshToken(_ context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error)
	CreateNewCredential(auth *entity.Auth) (*proto.Credential, error)
}

func NewService(
	repo auth_repo.Repository,
	chulaSSOClient chula_sso.ChulaSSO,
	tokenService token_svc.Service,
	userService user_svc.Service,
	conf cfgldr.App,
) Service {
	return auth.NewService(repo, chulaSSOClient, tokenService, userService, conf)
}
