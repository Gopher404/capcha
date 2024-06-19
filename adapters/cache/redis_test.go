package cache

import (
	"capcha/config"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSetGetDelete(t *testing.T) {

	cnf, err := config.GetConfig("")
	require.NoError(t, err)
	r, err := NewRedisCache(cnf.Redis.Addr, cnf.Redis.Password)
	require.NoError(t, err, "cannot connect to redis server")

	const (
		key = "test_key"
		val = "test_val"
	)

	err = r.Set(key, val, time.Minute)
	require.NoError(t, err)
	v, err := r.Get(key)
	require.NoError(t, err)
	require.Equal(t, val, v)
	err = r.Delete(key)
	require.NoError(t, err)
}
