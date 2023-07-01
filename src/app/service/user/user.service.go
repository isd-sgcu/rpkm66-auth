package user

import (
	"context"
	"time"

	"github.com/isd-sgcu/rpkm66-auth/src/proto"
)

type serviceImpl struct {
	client proto.UserServiceClient
}

func NewUserService(client proto.UserServiceClient) *serviceImpl {
	return &serviceImpl{client: client}
}

func (s *serviceImpl) FindByStudentID(sid string) (*proto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5000)
	defer cancel()

	res, err := s.client.FindByStudentID(ctx, &proto.FindByStudentIDUserRequest{StudentId: sid})
	if err != nil {
		return nil, err
	}

	return res.User, nil
}

func (s *serviceImpl) Create(user *proto.User) (*proto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5000)
	defer cancel()

	res, err := s.client.Create(ctx, &proto.CreateUserRequest{User: user})
	if err != nil {
		return nil, err
	}

	return res.User, nil
}
