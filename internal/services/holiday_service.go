package services

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/cache"
	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
)

// HolidayService 节假日查询服务，带缓存
type HolidayService struct {
	repo  *repositories.HolidayRepository
	cache *cache.HolidayCache
}

func NewHolidayService(repo *repositories.HolidayRepository, cache *cache.HolidayCache) *HolidayService {
	return &HolidayService{repo: repo, cache: cache}
}

// GetHolidaysByYear 返回指定年份的节假日集合（dateKey → isHoliday），优先读缓存
func (s *HolidayService) GetHolidaysByYear(ctx context.Context, year int) (map[string]bool, error) {
	if m, ok := s.cache.Get(year); ok {
		return m, nil
	}
	holidays, err := s.repo.GetByYear(ctx, year)
	if err != nil {
		return nil, fmt.Errorf("查询节假日失败 year=%d: %w", year, err)
	}
	s.cache.Set(year, holidays)
	m, _ := s.cache.Get(year)
	return m, nil
}

// IsHoliday 判断指定日期是否为法定节假日（非调班补班日）
func (s *HolidayService) IsHoliday(ctx context.Context, date time.Time) (bool, error) {
	key := date.Format("2006-01-02")
	m, err := s.GetHolidaysByYear(ctx, date.Year())
	if err != nil {
		return false, err
	}
	return m[key], nil
}

// IsWorkday 判断指定日期是否为需要上班的工作日（含调班补班的周末）
func (s *HolidayService) IsWorkday(ctx context.Context, date time.Time) (bool, error) {
	key := date.Format("2006-01-02")
	m, err := s.GetHolidaysByYear(ctx, date.Year())
	if err != nil {
		return false, err
	}
	wd := date.Weekday()
	isWeekend := wd == time.Saturday || wd == time.Sunday
	isHoliday := m[key]
	// 调班补班的周末（is_workday=true）不在节假日集合中（Set 方法中已排除）
	// 正常工作日：非周末 且 非节假日；或者 虽是周末但标记为补班
	if isWeekend {
		// 周末默认不上班，仅在补班日集合中出现时才算工作日
		return !isHoliday && s.isCompensatoryWorkday(m, key), nil
	}
	return !isHoliday, nil
}

// isCompensatoryWorkday 检查该日期是否被标注为调班补班工作日
// 补班日在 Set 时 isWorkday=true 的条目会以 value=false 写入 map，
// 普通节假日以 value=true 写入，因此 key 存在且 value=false 代表补班
func (s *HolidayService) isCompensatoryWorkday(m map[string]bool, key string) bool {
	val, exists := m[key]
	return exists && !val
}

// BatchImport 批量导入节假日数据并刷新缓存（供 HR 管理接口调用）
func (s *HolidayService) BatchImport(ctx context.Context, holidays []*models.Holiday) error {
	if err := s.repo.BatchUpsert(ctx, holidays); err != nil {
		return fmt.Errorf("节假日导入失败: %w", err)
	}
	// 按年份分组刷新缓存
	years := make(map[int]bool)
	for _, h := range holidays {
		years[h.Year] = true
	}
	for year := range years {
		freshData, err := s.repo.GetByYear(ctx, year)
		if err != nil {
			continue
		}
		s.cache.Set(year, freshData)
	}
	return nil
}
