package auth

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/isd-sgcu/rpkm66-auth/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/client"
	role "github.com/isd-sgcu/rpkm66-auth/constant/auth"
	dto "github.com/isd-sgcu/rpkm66-auth/internal/dto/auth"
	entity "github.com/isd-sgcu/rpkm66-auth/internal/entity/auth"
	auth_proto "github.com/isd-sgcu/rpkm66-auth/internal/proto/rpkm66/auth/auth/v1"
	"github.com/isd-sgcu/rpkm66-auth/internal/utils"
	"github.com/isd-sgcu/rpkm66-auth/pkg/client/chula_sso"
	auth_repo "github.com/isd-sgcu/rpkm66-auth/pkg/repository/auth"
	token_svc "github.com/isd-sgcu/rpkm66-auth/pkg/service/token"
	user_svc "github.com/isd-sgcu/rpkm66-auth/pkg/service/user"
	user_proto "github.com/isd-sgcu/rpkm66-go-proto/rpkm66/backend/user/v1"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ auth_proto.AuthServiceServer = &serviceImpl{}

type serviceImpl struct {
	auth_proto.UnimplementedAuthServiceServer
	repo              auth_repo.Repository
	chulaSSOClient    chula_sso.ChulaSSO
	tokenService      token_svc.Service
	userService       user_svc.Service
	conf              cfgldr.App
	oauthConfig       *oauth2.Config
	googleOauthClient *client.GoogleOauthClient
}

func NewService(
	repo auth_repo.Repository,
	chulaSSOClient chula_sso.ChulaSSO,
	tokenService token_svc.Service,
	userService user_svc.Service,
	conf cfgldr.App,
	oauthConfig *oauth2.Config,
	googleOauthClient *client.GoogleOauthClient,
) *serviceImpl {
	return &serviceImpl{
		repo:              repo,
		chulaSSOClient:    chulaSSOClient,
		tokenService:      tokenService,
		userService:       userService,
		conf:              conf,
		oauthConfig:       oauthConfig,
		googleOauthClient: googleOauthClient,
	}
}

func (s *serviceImpl) VerifyTicket(_ context.Context, req *auth_proto.VerifyTicketRequest) (res *auth_proto.VerifyTicketResponse, err error) {
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

				in := &user_proto.User{
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

	return &auth_proto.VerifyTicketResponse{Credential: credentials}, err
}

func (s *serviceImpl) Validate(_ context.Context, req *auth_proto.ValidateRequest) (res *auth_proto.ValidateResponse, err error) {
	credential, err := s.tokenService.Validate(req.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &auth_proto.ValidateResponse{
		UserId: credential.UserId,
		Role:   string(credential.Role),
	}, nil
}

func (s *serviceImpl) RefreshToken(_ context.Context, req *auth_proto.RefreshTokenRequest) (res *auth_proto.RefreshTokenResponse, err error) {
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

	return &auth_proto.RefreshTokenResponse{Credential: credentials}, nil
}

func (s *serviceImpl) CreateNewCredential(auth *entity.Auth) (*auth_proto.Credential, error) {
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

func (s *serviceImpl) GetGoogleLoginUrl(context.Context, *auth_proto.GetGoogleLoginUrlRequest) (*auth_proto.GetGoogleLoginUrlResponse, error) {
	URL, err := url.Parse(s.oauthConfig.Endpoint.AuthURL)
	if err != nil {
		log.Error().Err(err).Msg("unable to parse url")
		return nil, status.Error(codes.Internal, "Internal server error")
	}
	parameters := url.Values{}
	parameters.Add("client_id", s.oauthConfig.ClientID)
	parameters.Add("scope", strings.Join(s.oauthConfig.Scopes, " "))
	parameters.Add("redirect_uri", s.oauthConfig.RedirectURL)
	parameters.Add("response_type", "code")
	URL.RawQuery = parameters.Encode()
	url := URL.String()

	return &auth_proto.GetGoogleLoginUrlResponse{
		Url: url,
	}, nil
}

func (s *serviceImpl) VerifyGoogleLogin(ctx context.Context, req *auth_proto.VerifyGoogleLoginRequest) (*auth_proto.VerifyGoogleLoginResponse, error) {
	code := req.GetCode()
	auth := entity.Auth{}

	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "No code is provided")
	}

	response, err := s.googleOauthClient.GetUserEmail(code)
	if err != nil {
		switch err.Error() {
		case "Invalid code":
			return nil, status.Error(codes.InvalidArgument, "Invalid code")
		default:
			log.Error().Err(err).Msg("Unable to get user info")
			return nil, status.Error(codes.Internal, "Internal server error")
		}
	}

	email := response.Email

	log.Info().Interface("response", response).Msg("kuy")

	ouid, err := utils.GetOuidFromGmail(email)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "Only chula student can login")
	}
	firstName := response.Firstname
	familyName := response.Lastname

	user, err := s.userService.FindByStudentID(ouid)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				year, err := utils.CalYearFromID(ouid)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "google").
						Str("student_id", ouid).
						Msg("Cannot parse year to to int")
					return nil, status.Error(codes.Internal, "Internal service error")
				}

				yearInt, err := strconv.Atoi(year)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "google").
						Str("student_id", ouid).
						Msg("Cannot parse student id to int")
					return nil, status.Error(codes.Internal, "Internal service error")
				}

				if yearInt > s.conf.MaxRestrictYear {
					log.Error().
						Str("service", "auth").
						Str("module", "google").
						Str("student_id", ouid).
						Msg("Someone is trying to login (forbidden year)")
					return nil, status.Error(codes.PermissionDenied, "Forbidden study year")
				}

				faculty, err := utils.GetFacultyFromID(ouid)
				if err != nil {
					log.Error().
						Err(err).
						Str("service", "auth").
						Str("module", "google").
						Str("student_id", ouid).
						Msg("Cannot get faculty from student id")
					return nil, status.Error(codes.Internal, "Internal service error")
				}

				in := &user_proto.User{
					Firstname: firstName,
					Lastname:  familyName,
					StudentID: ouid,
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
						Str("module", "google").
						Str("student_id", ouid).
						Msg("Error creating the auth data")
					return nil, status.Error(codes.Unavailable, st.Message())
				}

			default:
				log.Error().
					Err(err).
					Str("service", "auth").
					Str("module", "google").
					Str("student_id", ouid).
					Msg("Service is down")
				return nil, status.Error(codes.Unavailable, st.Message())
			}
		} else {
			log.Error().
				Err(err).
				Str("service", "auth").
				Str("module", "google").
				Str("student_id", ouid).
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

	return &auth_proto.VerifyGoogleLoginResponse{Credential: credentials}, err
}
