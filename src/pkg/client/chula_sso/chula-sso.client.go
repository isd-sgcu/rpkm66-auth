package chula_sso

import (
	"github.com/isd-sgcu/rpkm66-auth/src/app/dto/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/src/client"
)

type ChulaSSO interface {
	VerifyTicket(ticket string, result *auth.ChulaSSOCredential) error
}

func NewChulaSSO(conf cfgldr.ChulaSSO) ChulaSSO {
	return client.NewChulaSSO(conf)
}
