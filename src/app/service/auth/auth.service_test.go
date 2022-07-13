package auth

import (
	"context"
	"github.com/bxcodec/faker/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	dto "github.com/isd-sgcu/rnkm65-auth/src/app/dto/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/app/model"
	"github.com/isd-sgcu/rnkm65-auth/src/app/model/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/app/utils"
	role "github.com/isd-sgcu/rnkm65-auth/src/constant/auth"
	mock "github.com/isd-sgcu/rnkm65-auth/src/mocks/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"testing"
	"time"
)

type AuthServiceTest struct {
	suite.Suite
	Auth            *auth.Auth
	UserDto         *proto.User
	Credential      *proto.Credential
	Payload         *dto.TokenPayloadAuth
	UserCredential  *dto.UserCredential
	secret          string
	UnauthorizedErr error
	NotFoundErr     error
	ServiceDownErr  error
}

func TestAuthService(t *testing.T) {
	suite.Run(t, new(AuthServiceTest))
}

func (t *AuthServiceTest) SetupTest() {
	t.Auth = &auth.Auth{
		Base: model.Base{
			ID:        uuid.New(),
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			DeletedAt: gorm.DeletedAt{},
		},
		UserID:       faker.UUIDDigit(),
		Role:         role.USER,
		RefreshToken: faker.Word(),
	}

	t.UserDto = &proto.User{
		Id:              t.Auth.UserID,
		Firstname:       faker.FirstName(),
		Lastname:        faker.LastName(),
		Nickname:        faker.Name(),
		StudentID:       "63xxxxxx21",
		Faculty:         "Faculty of Engineering",
		Year:            "3",
		Title:           faker.Word(),
		Phone:           faker.Phonenumber(),
		LineID:          faker.Word(),
		Email:           faker.Email(),
		AllergyFood:     faker.Word(),
		FoodRestriction: faker.Word(),
		AllergyMedicine: faker.Word(),
		Disease:         faker.Word(),
		CanSelectBaan:   true,
	}

	t.Credential = &proto.Credential{
		AccessToken:  faker.Word(),
		RefreshToken: t.Auth.RefreshToken,
		ExpiresIn:    3600,
	}

	t.Payload = &dto.TokenPayloadAuth{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    faker.Word(),
			ExpiresAt: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserId: t.Auth.UserID,
	}

	t.UserCredential = &dto.UserCredential{
		UserId: t.Auth.UserID,
		Role:   role.Role(t.Auth.Role),
	}

	t.UnauthorizedErr = errors.New("unauthorized")
	t.NotFoundErr = errors.New("not found user")
	t.ServiceDownErr = errors.New("service is down")

	t.secret = "asuperstrong32bitpasswordgohere!"
}

func (t *AuthServiceTest) TestVerifyTicketSuccessFirstTimeLogin() {
	want := &proto.VerifyTicketResponse{
		Credential: t.Credential,
	}

	t.Auth.RefreshToken = utils.Hash([]byte(t.Auth.RefreshToken))

	ticket := faker.Word()
	chulaSSORes := &dto.ChulaSSOCredential{
		UID:         faker.Word(),
		Username:    faker.Username(),
		Gecos:       faker.Username(),
		Email:       faker.Email(),
		Disable:     false,
		Roles:       []string{"student"},
		Firstname:   t.UserDto.Firstname,
		Lastname:    t.UserDto.Lastname,
		FirstnameTH: faker.FirstName(),
		LastnameTH:  faker.LastName(),
		Ouid:        t.UserDto.StudentID,
	}

	a := &auth.Auth{
		UserID: t.UserDto.Id,
		Role:   role.USER,
	}

	repo := &mock.RepositoryMock{}
	repo.On("Create", a).Return(t.Auth, nil)
	repo.On("Update", t.Auth).Return(t.Auth, nil)

	chulaSSOClient := &mock.ChulaSSOClientMock{}
	chulaSSOClient.On("VerifyTicket", ticket, &dto.ChulaSSOCredential{}).Return(chulaSSORes, nil)

	in := &proto.User{
		StudentID: t.UserDto.StudentID,
		Faculty:   t.UserDto.Faculty,
		Year:      t.UserDto.Year,
		Firstname: t.UserDto.Firstname,
		Lastname:  t.UserDto.Lastname,
	}

	userService := &mock.UserServiceMock{}
	userService.On("FindByStudentID", t.UserDto.StudentID).Return(nil, status.Error(codes.NotFound, "not found user"))
	userService.On("Create", in).Return(t.UserDto, nil)

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("CreateCredentials", t.Auth, t.secret).Return(t.Credential, nil)

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)
	actual, err := srv.VerifyTicket(context.Background(), &proto.VerifyTicketRequest{Ticket: ticket})

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want, actual)
}

func (t *AuthServiceTest) TestVerifyTicketSuccessNotFirstTimeLogin() {
	ticket := faker.Word()

	repo := &mock.RepositoryMock{}
	repo.On("FindByUserID", t.UserDto.Id).Return(t.Auth, nil)
	repo.On("Update", t.Auth).Return(t.Auth, nil)

	chulaSSOClient := &mock.ChulaSSOClientMock{}
	chulaSSOClient.On("VerifyTicket", ticket, &dto.ChulaSSOCredential{}).Return(nil, errors.New("Invalid Ticket"))

	userService := &mock.UserServiceMock{}
	userService.On("FindByStudentID", t.UserDto.Id).Return(t.UserDto, nil)

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("CreateCredentials", t.Auth, t.secret).Return(t.Credential, nil)

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)
	actual, err := srv.VerifyTicket(context.Background(), &proto.VerifyTicketRequest{Ticket: ticket})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unauthenticated, st.Code())
}

func (t *AuthServiceTest) TestVerifyTicketInvalid() {
	ticket := faker.Word()

	repo := &mock.RepositoryMock{}

	chulaSSOClient := &mock.ChulaSSOClientMock{}
	chulaSSOClient.On("VerifyTicket", ticket, &dto.ChulaSSOCredential{}).Return(nil, errors.New("Invalid Ticket"))

	userService := &mock.UserServiceMock{}
	userService.On("FindByStudentID", t.UserDto.Id).Return(nil, t.NotFoundErr)

	tokenService := &mock.TokenServiceMock{}

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)
	actual, err := srv.VerifyTicket(context.Background(), &proto.VerifyTicketRequest{Ticket: ticket})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unauthenticated, st.Code())
}

func (t *AuthServiceTest) TestVerifyTicketGrpcErr() {
	ticket := faker.Word()

	repo := &mock.RepositoryMock{}

	chulaSSOClient := &mock.ChulaSSOClientMock{}
	chulaSSOClient.On("VerifyTicket", ticket, &dto.ChulaSSOCredential{}).Return(nil, errors.New("Invalid Ticket"))

	userService := &mock.UserServiceMock{}
	userService.On("FindByStudentID", t.UserDto.Id).Return(nil, t.NotFoundErr)

	tokenService := &mock.TokenServiceMock{}

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)
	actual, err := srv.VerifyTicket(context.Background(), &proto.VerifyTicketRequest{Ticket: ticket})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unauthenticated, st.Code())
}

func (t *AuthServiceTest) TestValidateSuccess() {
	want := &proto.ValidateResponse{
		UserId: t.UserDto.Id,
		Role:   t.Auth.Role,
	}
	token := faker.Word()

	repo := &mock.RepositoryMock{}

	chulaSSOClient := &mock.ChulaSSOClientMock{}

	userService := &mock.UserServiceMock{}

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("Validate", token).Return(t.UserCredential, nil)

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)

	actual, err := srv.Validate(context.Background(), &proto.ValidateRequest{Token: token})

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want, actual)
}

func (t *AuthServiceTest) TestValidateInvalidToken() {
	token := faker.Word()

	repo := &mock.RepositoryMock{}

	chulaSSOClient := &mock.ChulaSSOClientMock{}

	userService := &mock.UserServiceMock{}

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("Validate", token).Return(nil, errors.New("Invalid token"))

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)

	actual, err := srv.Validate(context.Background(), &proto.ValidateRequest{Token: token})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unauthenticated, st.Code())
}

func (t *AuthServiceTest) TestRedeemRefreshTokenSuccess() {
	token := faker.Word()
	t.Auth.RefreshToken = utils.Hash([]byte(t.Credential.RefreshToken))

	want := &proto.RefreshTokenResponse{Credential: t.Credential}

	repo := &mock.RepositoryMock{}
	repo.On("FindByRefreshToken", utils.Hash([]byte(token)), &auth.Auth{}).Return(t.Auth, nil)
	repo.On("Update", t.Auth).Return(t.Auth, nil)

	chulaSSOClient := &mock.ChulaSSOClientMock{}

	userService := &mock.UserServiceMock{}

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("CreateRefreshToken").Return(token)
	tokenService.On("CreateCredentials", t.Auth, t.secret).Return(t.Credential, nil)

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)

	actual, err := srv.RefreshToken(context.Background(), &proto.RefreshTokenRequest{RefreshToken: token})

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want, actual)
}

func (t *AuthServiceTest) TestRedeemRefreshTokenInvalidToken() {
	token := faker.Word()
	t.Credential.RefreshToken = utils.Hash([]byte(token))

	repo := &mock.RepositoryMock{}
	repo.On("FindByRefreshToken", t.Credential.RefreshToken, &auth.Auth{}).Return(nil, errors.New("Not found token"))
	repo.On("Update", t.Auth).Return(t.Auth, nil)

	chulaSSOClient := &mock.ChulaSSOClientMock{}

	userService := &mock.UserServiceMock{}

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("CreateRefreshToken").Return(token)
	tokenService.On("CreateCredentials", t.Auth, t.secret).Return(t.Credential, nil)

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)

	actual, err := srv.RefreshToken(context.Background(), &proto.RefreshTokenRequest{RefreshToken: token})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unauthenticated, st.Code())
}

func (t *AuthServiceTest) TestRedeemRefreshTokenInternalErr() {
	token := faker.Word()
	t.Credential.RefreshToken = utils.Hash([]byte(token))

	repo := &mock.RepositoryMock{}
	repo.On("FindByRefreshToken", t.Credential.RefreshToken, &auth.Auth{}).Return(t.Auth, nil)

	chulaSSOClient := &mock.ChulaSSOClientMock{}

	userService := &mock.UserServiceMock{}

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("CreateRefreshToken").Return(token)
	tokenService.On("CreateCredentials", t.Auth, t.secret).Return(nil, errors.New("Invalid secret key"))

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)

	actual, err := srv.RefreshToken(context.Background(), &proto.RefreshTokenRequest{RefreshToken: token})

	st, ok := status.FromError(err)

	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Internal, st.Code())
}

func (t *AuthServiceTest) TestCreateCredentialsSuccess() {
	token := faker.Word()
	t.Credential.RefreshToken = utils.Hash([]byte(faker.Word()))

	want := t.Credential

	repo := &mock.RepositoryMock{}
	repo.On("Update", t.Auth).Return(t.Auth, nil)

	chulaSSOClient := &mock.ChulaSSOClientMock{}

	userService := &mock.UserServiceMock{}

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("CreateRefreshToken").Return(token)
	tokenService.On("CreateCredentials", t.Auth, t.secret).Return(t.Credential, nil)

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)

	credentials, err := srv.CreateNewCredential(t.Auth)

	assert.Nilf(t.T(), err, "error: %v", err)
	assert.Equal(t.T(), want, credentials)
}

func (t *AuthServiceTest) TestCreateCredentialsInternalErr() {
	want := errors.New("Invalid secret key")

	token := faker.Word()
	t.Credential.RefreshToken = utils.Hash([]byte(faker.Word()))

	repo := &mock.RepositoryMock{}
	repo.On("Update", t.Auth).Return(t.Auth, nil)

	chulaSSOClient := &mock.ChulaSSOClientMock{}

	userService := &mock.UserServiceMock{}

	tokenService := &mock.TokenServiceMock{}
	tokenService.On("CreateRefreshToken").Return(token)
	tokenService.On("CreateCredentials", t.Auth, t.secret).Return(nil, errors.New("Invalid secret key"))

	srv := NewService(repo, chulaSSOClient, tokenService, userService, t.secret)

	credentials, err := srv.CreateNewCredential(t.Auth)

	assert.Nil(t.T(), credentials)
	assert.Equal(t.T(), want.Error(), err.Error())
}
