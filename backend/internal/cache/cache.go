package cache

import (
	"context"
	"encoding/json"
	"rag-searchbot-backend/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	Cache       map[string]interface{}
	RedisClient *redis.Client
	RedisTTL    time.Duration
}

// ใช้ prefix เพื่อแยก key ของ user cache
func getUserKey(email string) string {
	return "cache:user:" + email
}

// ลบ key ออกจาก memory และ Redis
func (s *Service) Delete(key string) {
	if s.Cache != nil {
		delete(s.Cache, key)
	}
	if s.RedisClient != nil {
		s.RedisClient.Del(context.Background(), key)
	}
}

// ใช้กับ user โดยระบุ email
func (s *Service) ClearUserCache(email string) {
	s.Delete(getUserKey(email))
}

// Set user -> JSON และเก็บทั้ง Redis และ memory
func (s *Service) SetUserCache(email string, user interface{}) error {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.Set(context.Background(), getUserKey(email), jsonData)
}

// Get user -> ดึงจาก Redis หรือ memory แล้วแปลงเป็น struct
func (s *Service) GetUserCache(email string) (*models.User, error) {
	val, exists := s.GetString(context.Background(), getUserKey(email))
	if !exists {
		return nil, nil
	}

	var user models.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// รองรับเก็บข้อมูลประเภท string หรือ []byte ได้
func (s *Service) Set(ctx context.Context, key string, value interface{}) error {
	s.Cache[key] = value

	if s.RedisClient != nil {
		var toStore string
		switch v := value.(type) {
		case []byte:
			toStore = string(v)
		case string:
			toStore = v
		default:
			jsonData, err := json.Marshal(v)
			if err != nil {
				return err
			}
			toStore = string(jsonData)
		}
		return s.RedisClient.Set(ctx, key, toStore, s.RedisTTL).Err()
	}
	return nil
}

// ดึงข้อมูลแบบ string จากทั้ง in-memory และ Redis
func (s *Service) GetString(ctx context.Context, key string) (string, bool) {
	val, ok := s.Get(ctx, key)
	if !ok {
		return "", false
	}

	switch v := val.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	default:
		return "", false
	}
}

// ดึงค่าจาก memory ก่อน → Redis fallback
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

// Clear ทั้ง cache memory (global)
func (s *Service) Clear() {
	s.Cache = make(map[string]interface{})
}
