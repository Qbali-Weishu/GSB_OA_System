package cache

import (
	"sync"
	"time"

	"github.com/company/oa-leave-system/internal/models"
)

// HolidayCache 节假日内存缓存，按年份存储，TTL 24小时
type HolidayCache struct {
	mu     sync.RWMutex
	data   map[int]map[string]bool // year → dateKey → isHoliday
	expiry map[int]time.Time
	ttl    time.Duration
}

func NewHolidayCache(ttl time.Duration) *HolidayCache {
	return &HolidayCache{
		data:   make(map[int]map[string]bool),
		expiry: make(map[int]time.Time),
		ttl:    ttl,
	}
}

// Get 读取指定年份的节假日集合，返回 false 表示缓存未命中或已过期
func (c *HolidayCache) Get(year int) (map[string]bool, bool) {
	c.mu.RLock()
	data, ok := c.data[year]
	exp, hasExp := c.expiry[year]
	c.mu.RUnlock()

	if !ok || !hasExp {
		return nil, false
	}
	if time.Now().After(exp) {
		// 缓存已过期，异步清理，本次返回 miss 触发重新加载
		go c.evict(year)
		return nil, false
	}
	return data, true
}

// Set 写入指定年份的节假日集合并更新过期时间
func (c *HolidayCache) Set(year int, holidays []*models.Holiday) {
	m := make(map[string]bool, len(holidays))
	for _, h := range holidays {
		key := h.Date.Format("2006-01-02")
		if h.IsWorkday {
			// 调班补班日：周末变工作日，从节假日集合中排除
			m[key] = false
		} else {
			m[key] = true
		}
	}
	c.mu.Lock()
	c.data[year] = m
	c.expiry[year] = time.Now().Add(c.ttl)
	c.mu.Unlock()
}

// evict 删除指定年份的缓存条目
func (c *HolidayCache) evict(year int) {
	c.mu.Lock()
	delete(c.data, year)
	delete(c.expiry, year)
	c.mu.Unlock()
}
