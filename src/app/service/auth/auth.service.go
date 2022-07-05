package auth

import (
	"context"
	dto "github.com/isd-sgcu/rnkm65-auth/src/app/dto/auth"
	model "github.com/isd-sgcu/rnkm65-auth/src/app/model/auth"
	"github.com/isd-sgcu/rnkm65-auth/src/app/utils"
	"github.com/isd-sgcu/rnkm65-auth/src/constant"
	"github.com/isd-sgcu/rnkm65-auth/src/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	repo           IRepository
	chulaSSOClient IChulaSSOClient
	tokenService   ITokenService
	userService    IUserService
	secret         string
}

type IRepository interface {
	FindByRefreshToken(string, *model.Auth) error
	FindByUserID(string, *model.Auth) error
	Create(*model.Auth) error
	Update(string, *model.Auth) error
}

type IChulaSSOClient interface {
	VerifyTicket(string, *dto.ChulaSSOCredential) error
}

type IUserService interface {
	FindByStudentID(string) (*proto.User, error)
	Create(*proto.User) (*proto.User, error)
}

type ITokenService interface {
	CreateCredentials(*model.Auth, string) (*proto.Credential, error)
	Validate(string) (*dto.TokenPayloadAuth, error)
}

func NewService(
	repo IRepository,
	chulaSSOClient IChulaSSOClient,
	tokenService ITokenService,
	userService IUserService,
	secret string,
) *Service {
	return &Service{
		repo:           repo,
		chulaSSOClient: chulaSSOClient,
		tokenService:   tokenService,
		userService:    userService,
		secret:         secret,
	}
}

func (s *Service) VerifyTicket(_ context.Context, req *proto.VerifyTicketRequest) (res *proto.VerifyTicketResponse, err error) {
	ssoData := dto.ChulaSSOCredential{}
	auth := model.Auth{}

	err = s.chulaSSOClient.VerifyTicket(req.Ticket, &ssoData)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	user, err := s.userService.FindByStudentID(ssoData.Ouid)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				year, err := utils.CalYearFromID(ssoData.Ouid)
				if err != nil {
					return nil, err
				}

				faculty, err := utils.GetFacultyFromID(ssoData.Ouid)
				if err != nil {
					return nil, err
				}

				in := &proto.User{
					StudentID: ssoData.Ouid,
					Year:      year,
					Faculty:   faculty.FacultyEN,
				}

				user, err = s.userService.Create(in)
				if err != nil {
					return nil, status.Error(codes.Unauthenticated, st.Message())
				}

				auth = model.Auth{
					Role:   constant.USER,
					UserID: user.Id,
				}

				err = s.repo.Create(&auth)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "verify ticket").
						Msg("Error creating the auth data")
					return nil, status.Error(codes.Unavailable, st.Message())
				}

			default:
				log.Error().
					Err(err).
					Str("service", "auth").
					Str("module", "verify ticket").
					Msg("Service is down")
				return nil, status.Error(codes.Unavailable, st.Message())
			}
		} else {
			log.Error().
				Err(err).
				Str("service", "auth").
				Str("module", "verify ticket").
				Msg("Error connect to sso")
			return nil, status.Error(codes.Unavailable, "Service is down")
		}
	} else {
		err := s.repo.FindByUserID(user.Id, &auth)
		if err != nil {
			return nil, status.Error(codes.NotFound, "not found user")
		}
	}

	credentials, err := s.CreateNewCredential(&auth)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.VerifyTicketResponse{Credential: credentials}, err
}

func (s *Service) Validate(_ context.Context, req *proto.ValidateRequest) (res *proto.ValidateResponse, err error) {
	payload, err := s.tokenService.Validate(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &proto.ValidateResponse{
		UserId: payload.UserId,
		Role:   payload.Role,
	}, nil
}

func (s *Service) RefreshToken(_ context.Context, req *proto.RefreshTokenRequest) (res *proto.RefreshTokenResponse, err error) {
	auth := model.Auth{}

	err = s.repo.FindByRefreshToken(req.RefreshToken, &auth)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid refresh token")
	}

	credentials, err := s.CreateNewCredential(&auth)
	if err != nil {
		log.Error().Err(err).
			Str("service", "auth").
			Str("module", "refresh token").
			Msg("Error while create new token")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.RefreshTokenResponse{Credential: credentials}, nil
}

func (s *Service) CreateNewCredential(auth *model.Auth) (*proto.Credential, error) {
	credentials, err := s.tokenService.CreateCredentials(auth, s.secret)
	if err != nil {
		return nil, err
	}

	auth.RefreshToken = credentials.RefreshToken

	err = s.repo.Update(auth.ID.String(), auth)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}
