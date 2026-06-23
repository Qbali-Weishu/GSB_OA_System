package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeaveBalanceRepository 假期余额数据访问层
type LeaveBalanceRepository struct {
	db *pgxpool.Pool
}

func NewLeaveBalanceRepository(db *pgxpool.Pool) *LeaveBalanceRepository {
	return &LeaveBalanceRepository{db: db}
}

// GetByUserAndType 查询指定用户、年份、假类的余额记录
func (r *LeaveBalanceRepository) GetByUserAndType(ctx context.Context, userID int64, year int, leaveType models.LeaveType) (*models.LeaveBalance, error) {
	b := &models.LeaveBalance{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, year, leave_type, total_days, used_days, pending_days, created_at, updated_at
		FROM leave_balances
		WHERE user_id=$1 AND year=$2 AND leave_type=$3`,
		userID, year, leaveType,
	).Scan(
		&b.ID, &b.UserID, &b.Year, &b.LeaveType,
		&b.TotalDays, &b.UsedDays, &b.PendingDays,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %w", err)
	}
	return b, nil
}

// Deduct 扣减余额（used_days 增加）
func (r *LeaveBalanceRepository) Deduct(ctx context.Context, userID int64, year int, leaveType models.LeaveType, days float64) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE leave_balances
		SET used_days = used_days + $1, updated_at = $2
		WHERE user_id=$3 AND year=$4 AND leave_type=$5
		  AND (total_days - used_days) >= $1`,
		days, time.Now(), userID, year, leaveType,
	)
	if err != nil {
		return fmt.Errorf("扣减余额失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("余额不足或记录不存在")
	}
	return nil
}

// Add 增加余额（撤销/拒绝时退还）
func (r *LeaveBalanceRepository) Add(ctx context.Context, userID int64, year int, leaveType models.LeaveType, days float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE leave_balances
		SET used_days = GREATEST(0, used_days - $1), updated_at = $2
		WHERE user_id=$3 AND year=$4 AND leave_type=$5`,
		days, time.Now(), userID, year, leaveType,
	)
	if err != nil {
		return fmt.Errorf("退还余额失败: %w", err)
	}
	return nil
}
