package redis

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestRedis(t *testing.T) {
	// TODO(gorexlv): add redis ci
	mr, err := miniredis.Run()
	if err != nil {
		t.Errorf("redis run failed:%v", err)
	}
	redisConfig := DefaultRedisConfig()
	redisConfig.Addrs = []string{mr.Addr()}
	redisConfig.Mode = StubMode
	redisClient := redisConfig.Build()
	pingErr := redisClient.Client.Ping().Err()
	if pingErr != nil {
		t.Errorf("redis ping failed:%v", pingErr)
	}
	st := redisClient.Stub().PoolStats()
	t.Logf("running status %+v", st)
	err = redisClient.Close()
	if err != nil {
		t.Errorf("redis close failed:%v", err)
	}
	st = redisClient.Stub().PoolStats()
	t.Logf("close status %+v", st)
}
