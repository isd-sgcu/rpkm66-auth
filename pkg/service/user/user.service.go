package user

import (
	user_svc "github.com/isd-sgcu/rpkm66-auth/internal/service/user"
	proto "github.com/isd-sgcu/rpkm66-go-proto/rpkm66/backend/user/v1"
)

type Service interface {
	FindByStudentID(sid string) (*proto.User, error)
	Create(user *proto.User) (*proto.User, error)
}

func NewUserService(client proto.UserServiceClient) Service {
	return user_svc.NewUserService(client)
}
