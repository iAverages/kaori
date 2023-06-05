package redis

import (
	"context"
	"kaori/internal/config"

	r "github.com/redis/go-redis/v9"
)

var rdb *r.Client

func Init(cfg *config.Config) {
	rdb = r.NewClient(&r.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDatabase,
	})
}

func GetLastToken() (string, error) {
	return rdb.Get(context.Background(), "token").Result()
}
