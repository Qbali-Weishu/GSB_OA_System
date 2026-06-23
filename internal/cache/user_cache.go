package cache

import (
	"sync"
	"time"

	"github.com/company/oa-leave-system/internal/models"
)

// UserCache 用户信息内存缓存，TTL 5分钟
// 主要用于减少认证中间件对数据库的重复查询
type UserCache struct {
	mu     sync.RWMutex
	data   map[int64]*userEntry
	ttl    time.Duration
}

type userEntry struct {
	user      *models.User
	expiresAt time.Time
}

func NewUserCache(ttl time.Duration) *UserCache {
	return &UserCache{
		data: make(map[int64]*userEntry),
		ttl:  ttl,
	}
}

// Get 读取用户信息缓存
func (c *UserCache) Get(userID int64) (*models.User, bool) {
	c.mu.RLock()
	entry, ok := c.data[userID]
	c.mu.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.user, true
}

// Set 写入用户信息缓存
func (c *UserCache) Set(user *models.User) {
	c.mu.Lock()
	c.data[user.ID] = &userEntry{
		user:      user,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

// Invalidate 主动失效指定用户的缓存（用户信息变更时调用）
func (c *UserCache) Invalidate(userID int64) {
	c.mu.Lock()
	delete(c.data, userID)
	c.mu.Unlock()
}
