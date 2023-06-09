package redis

import (
	"context"
	"encoding/json"
	"kaori/internal/common"
	"kaori/internal/config"
	"time"

	r "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const REDIS_TOKEN = "kaori:token"
const REDIS_PLAYING_CACHE = "kaori:playingnow"

type redisServierImpl struct {
	logger *zap.SugaredLogger
	config *config.Config
	rdb    *r.Client
}

type RedisService interface {
	GetLastToken() (*oauth2.Token, error)
	SaveToken(token []byte) error
	GetSongCache() (common.PlayingNow, error)
	SaveSongCache(song common.PlayingNow) error
}

func NewRedisService(logger *zap.SugaredLogger, config *config.Config) RedisService {
	logger.Info("Connecting to redis...", config.RedisHost+":"+config.RedisPort)
	rdb := r.NewClient(&r.Options{
		Addr:     config.RedisHost + ":" + config.RedisPort,
		Password: config.RedisPassword,
		DB:       config.RedisDatabase,
	})

	// Try to ping redis, so we fail fast if it's not available
	val := rdb.Ping(context.Background())
	if val.Err() != nil {
		logger.Fatal(val.Err())
	}
	logger.Info("Connected to redis")

	return &redisServierImpl{
		logger: logger,
		config: config,
		rdb:    rdb,
	}
}

func (s *redisServierImpl) GetLastToken() (*oauth2.Token, error) {
	res, err := s.rdb.Get(context.Background(), REDIS_TOKEN).Result()

	if err != nil {
		return nil, err
	}

	var token *oauth2.Token
	json.Unmarshal([]byte(res), &token)

	return token, nil
}

func (s *redisServierImpl) SaveToken(token []byte) error {
	return s.rdb.Set(context.Background(), REDIS_TOKEN, token, 0).Err()
}

func (s *redisServierImpl) GetSongCache() (common.PlayingNow, error) {
	s.logger.Info("Getting song cache from redis")
	res, err := s.rdb.Get(context.Background(), REDIS_PLAYING_CACHE).Result()
	if err != nil {
		return common.PlayingNow{}, err
	}

	var song common.PlayingNow
	json.Unmarshal([]byte(res), &song)

	return song, nil
}

func (s *redisServierImpl) SaveSongCache(song common.PlayingNow) error {
	s.logger.Info("Saving song cache to redis")
	json, err := json.Marshal(song)
	if err != nil {
		return err
	}
	return s.rdb.Set(context.Background(), REDIS_PLAYING_CACHE, json, time.Second).Err()
}
