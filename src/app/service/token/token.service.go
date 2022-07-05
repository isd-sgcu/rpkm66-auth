package token

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	dto "github.com/isd-sgcu/rnkm65-auth/src/app/dto/auth"
	model "github.com/isd-sgcu/rnkm65-auth/src/app/model/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/config"
	"github.com/isd-sgcu/rnkm65-auth/src/proto"
	"github.com/pkg/errors"
	"time"
)

type Service struct {
	jwtService IJwtService
}

type IJwtService interface {
	SignAuth(*model.Auth) (string, error)
	VerifyAuth(string) (*jwt.Token, error)
	GetConfig() *config.Jwt
}

func NewTokenService(jwtService IJwtService) *Service {
	return &Service{
		jwtService: jwtService,
	}
}

func (s *Service) CreateCredentials(auth *model.Auth, secret string) (*proto.Credential, error) {
	token, err := s.jwtService.SignAuth(auth)
	if err != nil {
		return nil, err
	}

	credential := &proto.Credential{
		AccessToken:  token,
		RefreshToken: s.CreateRefreshToken(),
		ExpiresIn:    s.jwtService.GetConfig().ExpiresIn,
	}

	return credential, nil
}

func (s *Service) Validate(token string) (*dto.TokenPayloadAuth, error) {
	t, err := s.jwtService.VerifyAuth(token)
	if err != nil {
		return nil, err
	}

	payload := t.Claims.(jwt.MapClaims)

	if payload["iss"] != s.jwtService.GetConfig().Issuer {
		return nil, errors.New("Invalid token")
	}

	if time.Unix(int64(payload["exp"].(float64)), 0).Before(time.Now()) {
		return nil, errors.New("Token is expired")
	}

	return &dto.TokenPayloadAuth{
		UserId: payload["user_id"].(string),
		Role:   payload["role"].(string),
	}, nil
}

func (s *Service) CreateRefreshToken() string {
	return uuid.New().String()
}
