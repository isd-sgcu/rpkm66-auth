package token

import (
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm66-auth/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/constant/auth"
	dto "github.com/isd-sgcu/rpkm66-auth/internal/dto/auth"
	base "github.com/isd-sgcu/rpkm66-auth/internal/entity"
	entity "github.com/isd-sgcu/rpkm66-auth/internal/entity/auth"
	auth_proto "github.com/isd-sgcu/rpkm66-auth/internal/proto/rpkm66/auth/auth/v1"
	mock "github.com/isd-sgcu/rpkm66-auth/mocks/auth"
	"github.com/isd-sgcu/rpkm66-auth/mocks/cache"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TokenServiceTest struct {
	suite.Suite
	Credential   *auth_proto.Credential
	Auth         *entity.Auth
	Token        *jwt.Token
	TokenDecoded jwt.MapClaims
	Payload      *dto.TokenPayloadAuth
	Conf         *cfgldr.Jwt
}

func TestTokenService(t *testing.T) {
	suite.Run(t, new(TokenServiceTest))
}

func (t *TokenServiceTest) SetupTest() {
	t.Conf = &cfgldr.Jwt{
		Secret:    faker.Word(),
		ExpiresIn: 3600,
		Issuer:    faker.Word(),
	}

	t.Credential = &auth_proto.Credential{
		AccessToken:  faker.Word(),
		RefreshToken: faker.Word(),
		ExpiresIn:    3600,
	}

	t.Auth = &entity.Auth{
		Base: base.Base{
			ID:        uuid.New(),
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
		UserID:       faker.UUIDDigit(),
		Role:         auth.USER,
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
	}

	t.TokenDecoded = jwt.MapClaims{}
	t.TokenDecoded["iss"] = t.Conf.Issuer
	t.TokenDecoded["iat"] = t.Token.Claims.(dto.TokenPayloadAuth).IssuedAt
	t.TokenDecoded["exp"] = float64(time.Now().Add(time.Second * time.Duration(t.Conf.ExpiresIn)).UnixNano())
	t.TokenDecoded["user_id"] = t.Auth.UserID
	t.TokenDecoded["role"] = t.Auth.Role
}

func (t *TokenServiceTest) TestCreateCredentialsSuccess() {
	want := t.Credential

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("SignAuth", t.Auth).Return(t.Credential.AccessToken, nil)
	jwtSrv.On("GetConfig").Return(t.Conf, nil)

	cacheData := &dto.CacheAuth{
		Token: t.Credential.AccessToken,
		Role:  auth.USER,
	}

	cacheRepo := cache.RepositoryMock{
		V: map[string]interface{}{},
	}
	cacheRepo.On("SaveCache", t.TokenDecoded["user_id"], cacheData, 3600).Return(nil)

	srv := NewService(&jwtSrv, &cacheRepo)

	actual, err := srv.CreateCredentials(t.Auth, "asuperstrong32bitpasswordgohere!")

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want.AccessToken, actual.AccessToken)
	assert.Equal(t.T(), want.ExpiresIn, actual.ExpiresIn)
}

func (t *TokenServiceTest) TestCreateCredentialsInternalErr() {
	want := errors.New("Error while signing the token")

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("SignAuth", t.Auth).Return("", errors.New("Error while signing the token"))

	cacheRepo := cache.RepositoryMock{}

	srv := NewService(&jwtSrv, &cacheRepo)

	actual, err := srv.CreateCredentials(t.Auth, "asuperstrong32bitpasswordgohere!")

	var credential *auth_proto.Credential

	assert.Equal(t.T(), credential, actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}

func (t *TokenServiceTest) TestValidateAccessTokenSuccess() {
	want := &dto.UserCredential{
		UserId: t.Token.Claims.(dto.TokenPayloadAuth).UserId,
		Role:   auth.Role(t.Auth.Role),
	}
	token := faker.Word()

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", token).Return(&jwt.Token{
		Claims: t.TokenDecoded,
		Valid:  true,
	}, nil)
	jwtSrv.On("GetConfig").Return(t.Conf, nil)

	cacheAuth := dto.CacheAuth{
		Token: token,
		Role:  auth.USER,
	}
	cacheRepo := cache.RepositoryMock{}
	cacheRepo.On("GetCache", t.TokenDecoded["user_id"], &dto.CacheAuth{}).Return(&cacheAuth, nil)

	srv := NewService(&jwtSrv, &cacheRepo)

	actual, err := srv.Validate(token)

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want, actual)
}

func (t *TokenServiceTest) TestValidateAccessTokenInvalidToken() {
	testValidateAccessTokenInvalidTokenMalformedToken(t.T(), faker.Word())

	t.TokenDecoded["iss"] = "something"

	testValidateAccessTokenInvalidTokenInvalidCase(t.T(), t.Conf, &jwt.Token{
		Claims: t.TokenDecoded,
		Valid:  true,
	}, "Invalid token")

	t.TokenDecoded["iss"] = t.Conf.Issuer
	t.TokenDecoded["exp"] = float64(time.Now().Unix())

	testValidateAccessTokenInvalidTokenInvalidCase(t.T(), t.Conf, &jwt.Token{
		Claims: t.TokenDecoded,
		Valid:  true,
	}, "Token is expired")
}

func testValidateAccessTokenInvalidTokenMalformedToken(t *testing.T, refreshToken string) {
	want := errors.New("Error while signing the token")

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", refreshToken).Return(nil, errors.New("Error while signing the token"))

	cacheRepo := cache.RepositoryMock{}

	srv := NewService(&jwtSrv, &cacheRepo)

	actual, err := srv.Validate(refreshToken)

	var payload *dto.UserCredential

	assert.Equal(t, payload, actual)
	assert.Equal(t, want.Error(), err.Error())
}

func testValidateAccessTokenInvalidTokenInvalidCase(t *testing.T, conf *cfgldr.Jwt, tokenDecoded *jwt.Token, errMsg string) {
	want := errors.New(errMsg)

	in := faker.Word()

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", in).Return(tokenDecoded, nil)
	jwtSrv.On("GetConfig").Return(conf, nil)

	cacheRepo := cache.RepositoryMock{}

	srv := NewService(&jwtSrv, &cacheRepo)

	actual, err := srv.Validate(in)

	var payload *dto.UserCredential

	assert.Equal(t, payload, actual)
	assert.Equal(t, want.Error(), err.Error())
}

func (t *TokenServiceTest) TestValidateAccessTokenNotMatchWithCache() {
	want := errors.New("Invalid token")
	token := faker.Word()

	cacheAuth := dto.CacheAuth{
		Token: faker.Word(),
		Role:  auth.Role(t.Auth.Role),
	}

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", token).Return(&jwt.Token{
		Claims: t.TokenDecoded,
		Valid:  true,
	}, nil)
	jwtSrv.On("GetConfig").Return(t.Conf, nil)

	cacheRepo := cache.RepositoryMock{}
	cacheRepo.On("GetCache", t.TokenDecoded["user_id"], &dto.CacheAuth{}).Return(&cacheAuth, nil)

	srv := NewService(&jwtSrv, &cacheRepo)

	actual, err := srv.Validate(token)

	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}

func (t *TokenServiceTest) TestValidateCacheNotFoundUser() {
	want := errors.New("Invalid token")
	token := faker.Word()

	jwtSrv := mock.JwtServiceMock{}
	jwtSrv.On("VerifyAuth", token).Return(&jwt.Token{
		Claims: t.TokenDecoded,
		Valid:  true,
	}, nil)
	jwtSrv.On("GetConfig").Return(t.Conf, nil)

	cacheRepo := cache.RepositoryMock{}
	cacheRepo.On("GetCache", t.TokenDecoded["user_id"], &dto.CacheAuth{}).Return(nil, redis.Nil)

	srv := NewService(&jwtSrv, &cacheRepo)

	actual, err := srv.Validate(token)

	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}
