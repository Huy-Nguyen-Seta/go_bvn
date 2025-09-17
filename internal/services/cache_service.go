package services

import (
	"blog-api/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	rdb *redis.Client
	ctx context.Context
}

func NewCacheService(rdb *redis.Client) *CacheService {
	return &CacheService{
		rdb: rdb,
		ctx: context.Background(),
	}
}

func (s *CacheService) GetPost(id int64) (*models.Post, error) {
	key := fmt.Sprintf("post:%d", id)
	
	val, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var post models.Post
	if err := json.Unmarshal([]byte(val), &post); err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *CacheService) SetPost(post *models.Post) error {
	key := fmt.Sprintf("post:%d", post.ID)
	
	data, err := json.Marshal(post)
	if err != nil {
		return err
	}

	// Set with 5 minutes TTL
	return s.rdb.Set(s.ctx, key, data, 5*time.Minute).Err()
}

func (s *CacheService) DeletePost(id int64) error {
	key := fmt.Sprintf("post:%d", id)
	return s.rdb.Del(s.ctx, key).Err()
}