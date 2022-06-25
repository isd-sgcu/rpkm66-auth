package client

import (
	"github.com/go-resty/resty/v2"
	"github.com/isd-sgcu/rnkm65-auth/src/app/dto/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/config"
	"github.com/pkg/errors"
	"net/http"
)

type ChulaSSO struct {
	client *resty.Client
}

func NewChulaSSO(conf config.ChulaSSO) *ChulaSSO {
	client := resty.New().
		SetHeader("DeeAppID", conf.DeeAppID).
		SetHeader("DeeAppSecret", conf.DeeAppSecret).
		SetBaseURL(conf.Host)

	return &ChulaSSO{client: client}
}

func (c *ChulaSSO) VerifyTicket(ticket string, result *auth.ChulaSSOCredential) error {
	res, err := c.client.R().
		SetHeader("DeeTicket", ticket).
		SetResult(&result).
		Post("/serviceValidation")

	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusOK {
		return errors.New("Invalid ticket")
	}

	return nil
}
