package chula_sso

import (
	"github.com/isd-sgcu/rpkm66-auth/src/app/dto/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/client"
	"github.com/isd-sgcu/rpkm66-auth/src/config"
)

type ChulaSSO interface {
	VerifyTicket(ticket string, result *auth.ChulaSSOCredential) error
}

func NewChulaSSO(conf config.ChulaSSO) ChulaSSO {
	return client.NewChulaSSO(conf)
}
