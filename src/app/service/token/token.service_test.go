package token

import (
	"github.com/bxcodec/faker/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	dto "github.com/isd-sgcu/rnkm65-auth/src/app/dto/auth"
	base "github.com/isd-sgcu/rnkm65-auth/src/app/model"
	model "github.com/isd-sgcu/rnkm65-auth/src/app/model/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/config"
	"github.com/isd-sgcu/rnkm65-auth/src/constant"
	mock "github.com/isd-sgcu/rnkm65-auth/src/mocks/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
)

type TokenServiceTest struct {
	suite.Suite
	Credential *proto.Credential
	Auth       *model.Auth
	Token      *jwt.Token
	Payload    *dto.TokenPayloadAuth
	Conf       *config.Jwt
}

func TestTokenService(t *testing.T) {
	suite.Run(t, new(TokenServiceTest))
}

func (t *TokenServiceTest) SetupTest() {
	t.Conf = &config.Jwt{
		Secret:    faker.Word(),
		ExpiresIn: 3600,
		Issuer:    faker.Word(),
	}

	t.Credential = &proto.Credential{
		AccessToken:  faker.Word(),
		RefreshToken: faker.Word(),
		ExpiresIn:    3600,
	}

	t.Auth = &model.Auth{
		Base: base.Base{
			ID:        uuid.New(),
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
		UserID:       faker.UUIDDigit(),
		Role:         constant.USER,
		RefreshToken: faker.Word(),
	}

	t.Token = &jwt.Token{
		Claims: dto.TokenPayloadAuth{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    t.Conf.Issuer,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(t.Conf.ExpiresIn))),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
			UserId: t.Auth.UserID,
			Role:   t.Auth.Role,
		},
		Valid: true,
	}

	t.Payload = &dto.TokenPayloadAuth{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.Conf.Issuer,
			ExpiresAt: t.Token.Claims.(dto.TokenPayloadAuth).ExpiresAt,
			IssuedAt:  t.Token.Claims.(dto.TokenPayloadAuth).IssuedAt,
		},
		UserId: t.Auth.UserID,
		Role:   t.Auth.Role,
	}
}

func (t *TokenServiceTest) TestCreateCredentialsSuccess() {
	want := t.Credential

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("SignAuth", t.Auth).Return(t.Credential.AccessToken, nil)
	jwtSrv.On("GetConfig").Return(t.Conf, nil)

	srv := NewTokenService(&jwtSrv)

	actual, err := srv.CreateCredentials(t.Auth, "asuperstrong32bitpasswordgohere!")

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want.AccessToken, actual.AccessToken)
	assert.Equal(t.T(), want.ExpiresIn, actual.ExpiresIn)
}

func (t *TokenServiceTest) TestCreateCredentialsInternalErr() {
	want := errors.New("Error while signing the token")

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("SignAuth", t.Auth).Return("", errors.New("Error while signing the token"))

	srv := NewTokenService(&jwtSrv)

	actual, err := srv.CreateCredentials(t.Auth, "asuperstrong32bitpasswordgohere!")

	var credential *proto.Credential

	assert.Equal(t.T(), credential, actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}

func (t *TokenServiceTest) TestValidateAccessTokenSuccess() {
	want := t.Payload
	token := faker.Word()

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", token).Return(t.Token, nil)
	jwtSrv.On("GetConfig").Return(t.Conf, nil)

	srv := NewTokenService(&jwtSrv)

	actual, err := srv.Validate(token)

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want, actual)
}

func (t *TokenServiceTest) TestValidateAccessTokenInvalidToken() {
	testValidateAccessTokenInvalidTokenMalformedToken(t.T(), faker.Word())
	testValidateAccessTokenInvalidTokenInvalidCase(t.T(), t.Conf, &jwt.Token{
		Claims: dto.TokenPayloadAuth{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    faker.Word(),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now()),
			},
			UserId: t.Auth.UserID,
			Role:   t.Auth.Role,
		},
		Valid: true,
	}, "Invalid token")
	testValidateAccessTokenInvalidTokenInvalidCase(t.T(), t.Conf, &jwt.Token{
		Claims: dto.TokenPayloadAuth{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    t.Conf.Issuer,
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now()),
			},
			UserId: t.Auth.UserID,
			Role:   t.Auth.Role,
		},
		Valid: true,
	}, "Token is expired")
}

func testValidateAccessTokenInvalidTokenMalformedToken(t *testing.T, refreshToken string) {
	want := errors.New("Error while signing the token")

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", refreshToken).Return(nil, errors.New("Error while signing the token"))

	srv := NewTokenService(&jwtSrv)

	actual, err := srv.Validate(refreshToken)

	var payload *dto.TokenPayloadAuth

	assert.Equal(t, payload, actual)
	assert.Equal(t, want.Error(), err.Error())
}

func testValidateAccessTokenInvalidTokenInvalidCase(t *testing.T, conf *config.Jwt, token *jwt.Token, errMsg string) {
	want := errors.New(errMsg)

	in := faker.Word()

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", in).Return(token, nil)
	jwtSrv.On("GetConfig").Return(conf, nil)

	srv := NewTokenService(&jwtSrv)

	actual, err := srv.Validate(in)

	var payload *dto.TokenPayloadAuth

	assert.Equal(t, payload, actual)
	assert.Equal(t, want.Error(), err.Error())
}
