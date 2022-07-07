package cache

import (
	dto "github.com/isd-sgcu/rnkm65-auth/src/app/dto/auth"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/mock"
)

type RepositoryMock struct {
	mock.Mock
	V map[string]interface{}
}

func (t *RepositoryMock) SaveCache(key string, v interface{}, ttl int) error {
	args := t.Called(key, v, ttl)

	log.Print(key)

	t.V[key] = v

	return args.Error(0)
}

func (t *RepositoryMock) GetCache(key string, v interface{}) error {
	args := t.Called(key, v)

	if args.Get(0) != nil {
		*v.(*dto.CacheAuth) = *args.Get(0).(*dto.CacheAuth)
	}

	return args.Error(1)
}
