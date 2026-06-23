package database

import (
	"context"
	"fmt"

	"github.com/company/oa-leave-system/internal/config"
	"github.com/redis/go-redis/v9"
)

// NewRedis 创建 Redis 客户端并验证连通性
func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("Redis 连通性检查失败: %w", err)
	}

	return client, nil
}
