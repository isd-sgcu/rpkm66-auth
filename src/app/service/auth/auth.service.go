package auth

import (
	"context"
	"strconv"

	dto "github.com/isd-sgcu/rpkm66-auth/src/app/dto/auth"
	entity "github.com/isd-sgcu/rpkm66-auth/src/app/entity/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/app/utils"
	"github.com/isd-sgcu/rpkm66-auth/src/cfgldr"
	role "github.com/isd-sgcu/rpkm66-auth/src/constant/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/pkg/client/chula_sso"
	auth_repo "github.com/isd-sgcu/rpkm66-auth/src/pkg/repository/auth"
	token_svc "github.com/isd-sgcu/rpkm66-auth/src/pkg/service/token"
	user_svc "github.com/isd-sgcu/rpkm66-auth/src/pkg/service/user"
	"github.com/isd-sgcu/rpkm66-auth/src/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serviceImpl struct {
	repo           auth_repo.Repository
	chulaSSOClient chula_sso.ChulaSSO
	tokenService   token_svc.Service
	userService    user_svc.Service
	conf           cfgldr.App
}

func NewService(
	repo auth_repo.Repository,
	chulaSSOClient chula_sso.ChulaSSO,
	tokenService token_svc.Service,
	userService user_svc.Service,
	conf cfgldr.App,
) *serviceImpl {
	return &serviceImpl{
		repo:           repo,
		chulaSSOClient: chulaSSOClient,
		tokenService:   tokenService,
		userService:    userService,
		conf:           conf,
	}
}

func (s *serviceImpl) VerifyTicket(_ context.Context, req *proto.VerifyTicketRequest) (res *proto.VerifyTicketResponse, err error) {
	ssoData := dto.ChulaSSOCredential{}
	auth := entity.Auth{}

	err = s.chulaSSOClient.VerifyTicket(req.Ticket, &ssoData)
	if err != nil {
		log.Error().
			Err(err).
			Str("service", "auth service").
			Str("module", "verify ticket").
			Msgf("Someone is trying to logging in using SSO ticket")
		return nil, err
	}

	user, err := s.userService.FindByStudentID(ssoData.Ouid)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				year, err := utils.CalYearFromID(ssoData.Ouid)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "verify ticket").
						Str("student_id", ssoData.Ouid).
						Msg("Cannot parse year to to int")
					return nil, status.Error(codes.Internal, "Internal service error")
				}

				yearInt, err := strconv.Atoi(year)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "verify ticket").
						Str("student_id", ssoData.Ouid).
						Msg("Cannot parse student id to int")
					return nil, status.Error(codes.Internal, "Internal service error")
				}

				if yearInt > s.conf.MaxRestrictYear {
					log.Error().
						Str("service", "auth").
						Str("module", "verify ticket").
						Str("student_id", ssoData.Ouid).
						Msg("Someone is trying to login (forbidden year)")
					return nil, status.Error(codes.PermissionDenied, "Forbidden study year")
				}

				faculty, err := utils.GetFacultyFromID(ssoData.Ouid)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "verify ticket").
						Str("student_id", ssoData.Ouid).
						Msg("Cannot get faculty from student id")
					return nil, status.Error(codes.Internal, "Internal service error")
				}

				in := &proto.User{
					Firstname: ssoData.Firstname,
					Lastname:  ssoData.Lastname,
					StudentID: ssoData.Ouid,
					Year:      year,
					Faculty:   faculty.FacultyEN,
				}

				user, err = s.userService.Create(in)
				if err != nil {
					return nil, status.Error(codes.InvalidArgument, st.Message())
				}

				auth = entity.Auth{
					Role:   role.USER,
					UserID: user.Id,
				}

				err = s.repo.Create(&auth)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "verify ticket").
						Str("student_id", ssoData.Ouid).
						Msg("Error creating the auth data")
					return nil, status.Error(codes.Unavailable, st.Message())
				}

			default:
				log.Error().
					Err(err).
					Str("service", "auth").
					Str("module", "verify ticket").
					Str("student_id", ssoData.Ouid).
					Msg("Service is down")
				return nil, status.Error(codes.Unavailable, st.Message())
			}
		} else {
			log.Error().
				Err(err).
				Str("service", "auth").
				Str("module", "verify ticket").
				Str("student_id", ssoData.Ouid).
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

	log.Info().
		Str("service", "auth").
		Str("module", "verify ticket").
		Str("student_id", user.StudentID).
		Msg("User login to the service")

	return &proto.VerifyTicketResponse{Credential: credentials}, err
}

func (s *serviceImpl) Validate(_ context.Context, req *proto.ValidateRequest) (res *proto.ValidateResponse, err error) {
	credential, err := s.tokenService.Validate(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &proto.ValidateResponse{
		UserId: credential.UserId,
		Role:   string(credential.Role),
	}, nil
}

func (s *serviceImpl) RefreshToken(_ context.Context, req *proto.RefreshTokenRequest) (res *proto.RefreshTokenResponse, err error) {
	auth := entity.Auth{}

	err = s.repo.FindByRefreshToken(utils.Hash([]byte(req.RefreshToken)), &auth)
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

func (s *serviceImpl) CreateNewCredential(auth *entity.Auth) (*proto.Credential, error) {
	credentials, err := s.tokenService.CreateCredentials(auth, s.conf.Secret)
	if err != nil {
		return nil, err
	}

	auth.RefreshToken = utils.Hash([]byte(credentials.RefreshToken))

	err = s.repo.Update(auth.ID.String(), auth)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}
