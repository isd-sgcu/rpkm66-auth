package user

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	mock "github.com/isd-sgcu/rpkm66-auth/mocks/user"
	user_proto "github.com/isd-sgcu/rpkm66-go-proto/rpkm66/backend/user/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceTest struct {
	suite.Suite
	UserDto         *user_proto.User
	UnauthorizedErr error
	NotFoundErr     error
	ServiceDownErr  error
}

func TestUserService(t *testing.T) {
	suite.Run(t, new(UserServiceTest))
}

func (t *UserServiceTest) SetupTest() {
	t.UserDto = &user_proto.User{
		Title:           faker.Word(),
		Firstname:       faker.FirstName(),
		Lastname:        faker.LastName(),
		Nickname:        faker.Name(),
		StudentID:       faker.Word(),
		Faculty:         faker.Word(),
		Year:            faker.Word(),
		Phone:           faker.Phonenumber(),
		LineID:          faker.Word(),
		Email:           faker.Email(),
		AllergyFood:     faker.Word(),
		FoodRestriction: faker.Word(),
		AllergyMedicine: faker.Word(),
		Disease:         faker.Word(),
		CanSelectBaan:   true,
	}

	t.UnauthorizedErr = errors.New("Unauthorized")
	t.NotFoundErr = errors.New("Not found user")
	t.ServiceDownErr = errors.New("Service is down")
}

func (t *UserServiceTest) TestFindByStudentIDSuccess() {
	want := t.UserDto

	c := &mock.ClientMock{}
	c.On("FindByStudentID", &user_proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(&user_proto.FindByStudentIDUserResponse{User: t.UserDto}, nil)

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *UserServiceTest) TestFindByStudentIDUnauthorized() {
	c := &mock.ClientMock{}
	c.On("FindByStudentID", &user_proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(nil, status.Error(codes.Unauthenticated, t.NotFoundErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	st, ok := status.FromError(err)
	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unauthenticated, st.Code())
}

func (t *UserServiceTest) TestFindByStudentIDNotFound() {
	c := &mock.ClientMock{}
	c.On("FindByStudentID", &user_proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(nil, status.Error(codes.NotFound, t.NotFoundErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	st, ok := status.FromError(err)
	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.NotFound, st.Code())
}

func (t *UserServiceTest) TestFindByStudentIDGrpcError() {
	c := &mock.ClientMock{}
	c.On("FindByStudentID", &user_proto.FindByStudentIDUserRequest{StudentId: t.UserDto.StudentID}).
		Return(nil, status.Error(codes.Unavailable, t.ServiceDownErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.FindByStudentID(t.UserDto.StudentID)

	st, ok := status.FromError(err)
	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}

func (t *UserServiceTest) TestCreateSuccess() {
	want := t.UserDto

	c := &mock.ClientMock{}
	c.On("Create", &user_proto.CreateUserRequest{User: &user_proto.User{}}).
		Return(&user_proto.CreateUserResponse{User: t.UserDto}, nil)

	srv := NewUserService(c)

	actual, err := srv.Create(&user_proto.User{})

	assert.Nil(t.T(), err)
	assert.Equal(t.T(), want, actual)
}

func (t *UserServiceTest) TestCreateGrpcErr() {
	c := &mock.ClientMock{}
	c.On("Create", &user_proto.CreateUserRequest{User: &user_proto.User{}}).
		Return(nil, status.Error(codes.Unavailable, t.ServiceDownErr.Error()))

	srv := NewUserService(c)

	actual, err := srv.Create(&user_proto.User{})

	st, ok := status.FromError(err)
	assert.True(t.T(), ok)
	assert.Nil(t.T(), actual)
	assert.Equal(t.T(), codes.Unavailable, st.Code())
}
