package client

import (
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/isd-sgcu/rpkm66-auth/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/internal/dto/auth"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChulaSSO struct {
	client *resty.Client
}

func NewChulaSSO(conf cfgldr.ChulaSSO) *ChulaSSO {
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
		log.Error().
			Str("service", "chula sso client").
			Str("module", "verify ticket").
			Str("student_id", result.Ouid).
			Str("ticket", ticket).
			Msg("Invalid ticket")
		return status.Error(codes.Unauthenticated, "Invalid ticket")
	}

	if res.StatusCode() == http.StatusTooManyRequests {
		log.Error().
			Str("service", "chula sso client").
			Str("module", "verify ticket").
			Str("student_id", result.Ouid).
			Msg("Reach SSO Limit")

		return status.Error(codes.ResourceExhausted, err.Error())
	}

	if res.StatusCode() != http.StatusOK {
		log.Error().
			Str("service", "chula sso client").
			Str("module", "verify ticket").
			Str("status", res.Status()).
			Str("body", string(res.Body())).
			Str("student_id", result.Ouid).
			Msg("Invalid sso status")

		return status.Error(codes.Unauthenticated, "Invalid ticket")
	}

	return nil
}
