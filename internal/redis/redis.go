package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"kaori/internal/config"

	r "github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

const REDIS_TOKEN = "kaori:token"

var rdb *r.Client

func Init(cfg *config.Config) {
	fmt.Println("Connecting to redis...", cfg.RedisHost+":"+cfg.RedisPort)
	rdb = r.NewClient(&r.Options{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDatabase,
	})
}

func GetLastToken() (*oauth2.Token, error) {
	if rdb == nil {
		return nil, fmt.Errorf("redis not initialized")
	}
	res, err := rdb.Get(context.Background(), REDIS_TOKEN).Result()

	if err != nil {
		return nil, err
	}

	fmt.Println("Token found in redis", res)

	var token *oauth2.Token
	json.Unmarshal([]byte(res), &token)

	return token, nil
}

func SaveToken(token []byte) error {
	if rdb == nil {
		return fmt.Errorf("redis not initialized")
	}
	return rdb.Set(context.Background(), REDIS_TOKEN, token, 0).Err()
}
