package user

import (
	"context"
	"time"

	user_proto "github.com/isd-sgcu/rpkm66-go-proto/rpkm66/backend/user/v1"
)

type serviceImpl struct {
	client user_proto.UserServiceClient
}

func NewUserService(client user_proto.UserServiceClient) *serviceImpl {
	return &serviceImpl{client: client}
}

func (s *serviceImpl) FindByStudentID(sid string) (*user_proto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5000)
	defer cancel()

	res, err := s.client.FindByStudentID(ctx, &user_proto.FindByStudentIDUserRequest{StudentId: sid})
	if err != nil {
		return nil, err
	}

	return res.User, nil
}

func (s *serviceImpl) Create(user *user_proto.User) (*user_proto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5000)
	defer cancel()

	res, err := s.client.Create(ctx, &user_proto.CreateUserRequest{User: user})
	if err != nil {
		return nil, err
	}

	return res.User, nil
}
