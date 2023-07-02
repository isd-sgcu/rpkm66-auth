package user

import (
	"context"

	user_proto "github.com/isd-sgcu/rpkm66-go-proto/rpkm66/backend/user/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type ClientMock struct {
	mock.Mock
}

func (c *ClientMock) FindByStudentID(_ context.Context, in *user_proto.FindByStudentIDUserRequest, _ ...grpc.CallOption) (res *user_proto.FindByStudentIDUserResponse, err error) {
	args := c.Called(in)

	if args.Get(0) != nil {
		res = args.Get(0).(*user_proto.FindByStudentIDUserResponse)
	}

	return res, args.Error(1)
}

func (c *ClientMock) Create(_ context.Context, in *user_proto.CreateUserRequest, _ ...grpc.CallOption) (res *user_proto.CreateUserResponse, err error) {
	args := c.Called(in)

	if args.Get(0) != nil {
		res = args.Get(0).(*user_proto.CreateUserResponse)
	}

	return res, args.Error(1)
}

// Unused
func (c *ClientMock) FindOne(ctx context.Context, in *user_proto.FindOneUserRequest, opts ...grpc.CallOption) (*user_proto.FindOneUserResponse, error) {
	return nil, nil
}

// Unused
func (c *ClientMock) Update(ctx context.Context, in *user_proto.UpdateUserRequest, opts ...grpc.CallOption) (*user_proto.UpdateUserResponse, error) {
	return nil, nil
}

// Unused
func (c *ClientMock) Verify(ctx context.Context, in *user_proto.VerifyUserRequest, opts ...grpc.CallOption) (*user_proto.VerifyUserResponse, error) {
	return nil, nil
}

// Unused
func (c *ClientMock) Delete(ctx context.Context, in *user_proto.DeleteUserRequest, opts ...grpc.CallOption) (*user_proto.DeleteUserResponse, error) {
	return nil, nil
}

// Unused
func (c *ClientMock) CreateOrUpdate(ctx context.Context, in *user_proto.CreateOrUpdateUserRequest, opts ...grpc.CallOption) (*user_proto.CreateOrUpdateUserResponse, error) {
	return nil, nil
}

// Unused
func (c *ClientMock) ConfirmEstamp(ctx context.Context, in *user_proto.ConfirmEstampRequest, opts ...grpc.CallOption) (*user_proto.ConfirmEstampResponse, error) {
	return nil, nil
}

// Unused
func (c *ClientMock) GetUserEstamp(ctx context.Context, in *user_proto.GetUserEstampRequest, opts ...grpc.CallOption) (*user_proto.GetUserEstampResponse, error) {
	return nil, nil
}
