package chula_sso

import (
	"github.com/isd-sgcu/rpkm66-auth/client"
	"golang.org/x/oauth2"
)

type GoogleOauthClient interface {
	GetUserEmail(code string) (*client.GoogleUserEmailResponse, error)
}

func NewGoogleOauthClient(conf *oauth2.Config) GoogleOauthClient {
	return client.NewGoogleOauthClient(conf)
}
