package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/forwardalex/ysocks/conf"
)

var (
	Client *redis.Client
	Nil    = redis.Nil
)

// Init 初始化连接
func Init(cfg *conf.RedisConfig) (err error) {
	Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password, // no password set
		DB:           cfg.DB,       // use default DB
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	_, err = Client.Ping().Result()
	if err != nil {
		return err
	}
	//防止redis有key
	Client.Del("ip")
	return nil
}

func Close() {
	_ = Client.Close()
}
