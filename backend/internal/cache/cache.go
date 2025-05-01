package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	Cache       map[string]interface{}
	RedisClient *redis.Client
	RedisTTL    time.Duration
}

func (s *Service) GetString(ctx context.Context, key string) (string, bool) {
	val, ok := s.Get(ctx, key)
	if !ok {
		return "", false
	}

	str, ok := val.(string)
	return str, ok
}

// NewService creates a new cache service with Redis fallback
func NewService(redisClient *redis.Client, ttl time.Duration) *Service {
	return &Service{
		Cache:       make(map[string]interface{}),
		RedisClient: redisClient,
		RedisTTL:    ttl,
	}
}

func (s *Service) Set(ctx context.Context, key string, value interface{}) error {
	s.Cache[key] = value
	if s.RedisClient != nil {
		return s.RedisClient.Set(ctx, key, value, s.RedisTTL).Err()
	}
	return nil
}

func (s *Service) Get(ctx context.Context, key string) (interface{}, bool) {
	if val, ok := s.Cache[key]; ok {
		return val, true
	}

	if s.RedisClient != nil {
		val, err := s.RedisClient.Get(ctx, key).Result()
		if err == nil {
			s.Cache[key] = val // cache locally
			return val, true
		}
	}
	return nil, false
}

func (s *Service) Delete(key string) {
	delete(s.Cache, key)
	if s.RedisClient != nil {
		s.RedisClient.Del(context.Background(), key)
	}
}

func (s *Service) Clear() {
	s.Cache = make(map[string]interface{})
}

func (s *Service) GetByKey(ctx context.Context, key string) (interface{}, bool) {
	if val, ok := s.Cache[key]; ok {
		return val, true
	}

	if s.RedisClient != nil {
		val, err := s.RedisClient.Get(ctx, key).Result()
		if err == nil {
			s.Cache[key] = val // cache locally
			return val, true
		}
	}
	return nil, false
}
