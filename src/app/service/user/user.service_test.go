package user

import (
	"github.com/bxcodec/faker/v3"
	mock "github.com/isd-sgcu/rnkm65-auth/src/mocks/user"
	"github.com/isd-sgcu/rnkm65-auth/src/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

type UserServiceTest struct {
	suite.Suite
	UserDto         *proto.User
	UnauthorizedErr error
	NotFoundErr     error
	ServiceDownErr  error
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTest))
}

func (t *UserServiceTest) SetupTest() {
	t.UserDto = &proto.User{
		Id:                    faker.UUIDDigit(),
		Firstname:             faker.FirstName(),
		Lastname:              faker.LastName(),
		Nickname:              faker.Name(),
		StudentID:             faker.Word(),
		Faculty:               faker.Word(),
		Year:                  faker.Word(),
		Phone:                 faker.Phonenumber(),
		LineID:                faker.Word(),
		Email:                 faker.Email(),
		AllergyFood:           faker.Word(),
		FoodRestriction:       faker.Word(),
		AllergyMedicine:       faker.Word(),
		Disease:               faker.Word(),
		VaccineCertificateUrl: faker.URL(),
		ImageUrl:              faker.URL(),
	}

	t.UnauthorizedErr = errors.New("unauthorized")
	t.NotFoundErr = errors.New("not found user")
	t.ServiceDownErr = errors.New("service is down")
}

func (t *UserServiceTest) TestFindByStudentIDSuccess() {
	want := t.UserDto

	c := &mock.ClientMock{}
	c.On("FindByStudentID", &proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(&proto.FindByStudentIDUserResponse{User: t.UserDto}, nil)

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *UserServiceTest) TestFindByStudentIDUnauthorized() {
	want := t.UnauthorizedErr

	c := &mock.ClientMock{}
	c.On("FindByStudentID", &proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(nil, status.Error(codes.Unauthenticated, t.UnauthorizedErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}

func (t *UserServiceTest) TestFindByStudentIDNotFound() {
	want := t.NotFoundErr

	c := &mock.ClientMock{}
	c.On("FindByStudentID", &proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(nil, status.Error(codes.NotFound, t.NotFoundErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}

func (t *UserServiceTest) TestFindByStudentIDGrpcError() {
	want := t.ServiceDownErr

	c := &mock.ClientMock{}
	c.On("FindByStudentID", &proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(nil, status.Error(codes.Unavailable, t.ServiceDownErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}

func (t *UserServiceTest) TestCreateSuccess() {
	want := t.UserDto

	c := &mock.ClientMock{}
	c.On("Create", &proto.CreateUserRequest{User: &proto.User{}}).
		Return(&proto.CreateUserResponse{User: t.UserDto}, nil)

	srv := NewUserService(c)

	actual, err := srv.Create(&proto.User{})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *UserServiceTest) TestCreateUnauthorized() {
	want := t.UnauthorizedErr

	c := &mock.ClientMock{}
	c.On("Create", &proto.CreateUserRequest{User: &proto.User{}}).
		Return(nil, status.Error(codes.Unauthenticated, t.UnauthorizedErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.Create(&proto.User{})

	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}

func (t *UserServiceTest) TestCreateGrpcErr() {
	want := t.ServiceDownErr

	c := &mock.ClientMock{}
	c.On("Create", &proto.CreateUserRequest{User: &proto.User{}}).
		Return(nil, status.Error(codes.Unavailable, t.ServiceDownErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.Create(&proto.User{})

	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), want.Error(), err.Error())
}
