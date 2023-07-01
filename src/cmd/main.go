package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/isd-sgcu/rpkm66-auth/src/cfgldr"
	"github.com/isd-sgcu/rpkm66-auth/src/client"
	"github.com/isd-sgcu/rpkm66-auth/src/database"
	ar "github.com/isd-sgcu/rpkm66-auth/src/pkg/repository/auth"
	"github.com/isd-sgcu/rpkm66-auth/src/pkg/repository/cache"
	as "github.com/isd-sgcu/rpkm66-auth/src/pkg/service/auth"
	js "github.com/isd-sgcu/rpkm66-auth/src/pkg/service/jwt"
	ts "github.com/isd-sgcu/rpkm66-auth/src/pkg/service/token"
	"github.com/isd-sgcu/rpkm66-auth/src/pkg/service/user"
	jsg "github.com/isd-sgcu/rpkm66-auth/src/pkg/strategy"
	"github.com/isd-sgcu/rpkm66-auth/src/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type operation func(ctx context.Context) error

func gracefulShutdown(ctx context.Context, timeout time.Duration, ops map[string]operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)

		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		sig := <-s

		log.Info().
			Str("service", "graceful shutdown").
			Msgf("got signal \"%v\" shutting down service", sig)

		timeoutFunc := time.AfterFunc(timeout, func() {
			log.Error().
				Str("service", "graceful shutdown").
				Msgf("timeout %v ms has been elapsed, force exit", timeout.Milliseconds())
			os.Exit(0)
		})

		defer timeoutFunc.Stop()

		var wg sync.WaitGroup

		for key, op := range ops {
			wg.Add(1)
			innerOp := op
			innerKey := key
			go func() {
				defer wg.Done()

				log.Info().
					Str("service", "graceful shutdown").
					Msgf("cleaning up: %v", innerKey)
				if err := innerOp(ctx); err != nil {
					log.Error().
						Str("service", "graceful shutdown").
						Err(err).
						Msgf("%v: clean up failed: %v", innerKey, err.Error())
					return
				}

				log.Info().
					Str("service", "graceful shutdown").
					Msgf("%v was shutdown gracefully", innerKey)
			}()
		}

		wg.Wait()
		close(wait)
	}()

	return wait
}

func main() {
	conf, err := cfgldr.LoadConfig()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "auth").
			Msg("Failed to start service")
	}

	db, err := database.InitDatabase(&conf.Database)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "auth").
			Msg("Failed to start service")
	}

	cacheDB, err := database.InitRedisConnect(&conf.Redis)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "auth").
			Msg("Failed to start service")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", conf.App.Port))
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "auth").
			Msg("Failed to start service")
	}

	backendConn, err := grpc.Dial(conf.Service.Backend, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().
			Err(err).
			Str("service", "rnkm-backend").
			Msg("Cannot connect to service")
	}

	grpcServer := grpc.NewServer()

	cSSO := client.NewChulaSSO(conf.ChulaSSO)

	cacheRepo := cache.NewRepository(cacheDB)

	usrClient := proto.NewUserServiceClient(backendConn)
	usrSrv := user.NewUserService(usrClient)

	stg := jsg.NewJwtStrategy(conf.Jwt.Secret)
	jtSrv := js.NewJwtService(conf.Jwt, stg)

	tkSrv := ts.NewTokenService(jtSrv, cacheRepo)

	aRepo := ar.NewRepository(db)
	aSrv := as.NewService(aRepo, cSSO, tkSrv, usrSrv, conf.App)

	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	proto.RegisterAuthServiceServer(grpcServer, aSrv)

	reflection.Register(grpcServer)
	go func() {
		log.Info().
			Str("service", "auth").
			Msgf("rpkm66 auth starting at port %v", conf.App.Port)

		if err = grpcServer.Serve(lis); err != nil {
			log.Fatal().
				Err(err).
				Str("service", "auth").
				Msg("Failed to start service")
		}
	}()

	wait := gracefulShutdown(context.Background(), 2*time.Second, map[string]operation{
		"database": func(ctx context.Context) error {
			sqlDb, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDb.Close()
		},
		"server": func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return nil
		},
		"cache": func(ctx context.Context) error {
			return cacheDB.Close()
		},
	})

	<-wait

	grpcServer.GracefulStop()
	log.Info().
		Str("service", "auth").
		Msg("Closing the listener")
	lis.Close()
	log.Info().
		Str("service", "auth").
		Msg("End of Program")
}
