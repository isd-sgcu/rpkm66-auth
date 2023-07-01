package chula_sso

import (
	"github.com/isd-sgcu/rpkm66-auth/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/client"
	"github.com/isd-sgcu/rpkm66-auth/internal/dto/auth"
)

type ChulaSSO interface {
	VerifyTicket(ticket string, result *auth.ChulaSSOCredential) error
}

func NewChulaSSO(conf cfgldr.ChulaSSO) ChulaSSO {
	return client.NewChulaSSO(conf)
}
