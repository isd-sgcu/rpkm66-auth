package user

import (
	user_svc "github.com/isd-sgcu/rpkm66-auth/src/app/service/user"
	"github.com/isd-sgcu/rpkm66-auth/src/proto"
)

type Service interface {
	FindByStudentID(sid string) (*proto.User, error)
	Create(user *proto.User) (*proto.User, error)
}

func NewUserService(client proto.UserServiceClient) Service {
	return user_svc.NewUserService(client)
}
