package auth

import (
	"github.com/isd-sgcu/rpkm66-auth/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/client"
	proto "github.com/isd-sgcu/rpkm66-auth/internal/proto/rpkm66/auth/auth/v1"
	"github.com/isd-sgcu/rpkm66-auth/internal/service/auth"
	"github.com/isd-sgcu/rpkm66-auth/pkg/client/chula_sso"
	auth_repo "github.com/isd-sgcu/rpkm66-auth/pkg/repository/auth"
	token_svc "github.com/isd-sgcu/rpkm66-auth/pkg/service/token"
	user_svc "github.com/isd-sgcu/rpkm66-auth/pkg/service/user"
	"golang.org/x/oauth2"
)

func NewService(
	repo auth_repo.Repository,
	chulaSSOClient chula_sso.ChulaSSO,
	tokenService token_svc.Service,
	userService user_svc.Service,
	conf cfgldr.App,
	oauth *oauth2.Config,
	googleOauthClient *client.GoogleOauthClient,
) proto.AuthServiceServer {
	return auth.NewService(repo, chulaSSOClient, tokenService, userService, conf, oauth, googleOauthClient)
}
