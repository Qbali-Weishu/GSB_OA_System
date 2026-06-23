package repositories

import (
	"context"
	"fmt"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HolidayRepository 节假日数据访问层
type HolidayRepository struct {
	db *pgxpool.Pool
}

func NewHolidayRepository(db *pgxpool.Pool) *HolidayRepository {
	return &HolidayRepository{db: db}
}

// GetByYear 查询指定年份的所有节假日，包含调班补班日
func (r *HolidayRepository) GetByYear(ctx context.Context, year int) ([]*models.Holiday, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, date, name, is_workday, description, year, created_at
		FROM holidays
		WHERE year = $1
		ORDER BY date`, year,
	)
	if err != nil {
		return nil, fmt.Errorf("查询节假日失败: %w", err)
	}
	defer rows.Close()

	var holidays []*models.Holiday
	for rows.Next() {
		h := &models.Holiday{}
		if err := rows.Scan(
			&h.ID, &h.Date, &h.Name, &h.IsWorkday, &h.Description, &h.Year, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		holidays = append(holidays, h)
	}
	return holidays, nil
}

// BatchUpsert 批量写入或更新节假日（HR 录入下一年度日历时调用）
func (r *HolidayRepository) BatchUpsert(ctx context.Context, holidays []*models.Holiday) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, h := range holidays {
		_, err := tx.Exec(ctx, `
			INSERT INTO holidays (date, name, is_workday, description, year, created_at)
			VALUES ($1,$2,$3,$4,$5,NOW())
			ON CONFLICT (date) DO UPDATE
			SET name=$2, is_workday=$3, description=$4, year=$5`,
			h.Date, h.Name, h.IsWorkday, h.Description, h.Year,
		)
		if err != nil {
			return fmt.Errorf("写入节假日记录失败 date=%s: %w", h.Date.Format("2006-01-02"), err)
		}
	}

	return tx.Commit(ctx)
}
