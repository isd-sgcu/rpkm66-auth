package user

import (
	"context"
	"github.com/isd-sgcu/rnkm65-auth/src/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
	"time"
)

type Service struct {
	client proto.UserServiceClient
}

func NewUserService(client proto.UserServiceClient) *Service {
	return &Service{client: client}
}

func (s *Service) FindByStudentID(sid string) (*proto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5000)
	defer cancel()

	res, err := s.client.FindByStudentID(ctx, &proto.FindByStudentIDUserRequest{StudentId: sid})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, errors.New(st.Message())
		}
		return nil, errors.New("Service is down")
	}

	return res.User, nil
}

func (s *Service) Create(user *proto.User) (*proto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5000)
	defer cancel()

	res, err := s.client.Create(ctx, &proto.CreateUserRequest{User: user})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, errors.New(st.Message())
		}
		return nil, errors.New("Service is down")
	}

	return res.User, nil
}
