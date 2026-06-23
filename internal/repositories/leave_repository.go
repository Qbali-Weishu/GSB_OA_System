package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LeaveRepository 请假单数据访问层
type LeaveRepository struct {
	db *pgxpool.Pool
}

func NewLeaveRepository(db *pgxpool.Pool) *LeaveRepository {
	return &LeaveRepository{db: db}
}

// Create 创建请假申请单
func (r *LeaveRepository) Create(ctx context.Context, leave *models.Leave) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO leaves (user_id, dept_id, leave_type, start_date, end_date,
		                    working_days, reason, status, required_approvals, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id`,
		leave.UserID, leave.DeptID, leave.LeaveType,
		leave.StartDate, leave.EndDate, leave.WorkingDays,
		leave.Reason, leave.Status, leave.RequiredApprovals,
		leave.CreatedAt, leave.UpdatedAt,
	).Scan(&leave.ID)
	if err != nil {
		return fmt.Errorf("创建请假记录失败: %w", err)
	}
	return nil
}

// GetByID 按主键查询请假单
func (r *LeaveRepository) GetByID(ctx context.Context, id int64) (*models.Leave, error) {
	leave := &models.Leave{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, dept_id, leave_type, start_date, end_date,
		       working_days, reason, status, required_approvals, created_at, updated_at
		FROM leaves WHERE id = $1`, id,
	).Scan(
		&leave.ID, &leave.UserID, &leave.DeptID, &leave.LeaveType,
		&leave.StartDate, &leave.EndDate, &leave.WorkingDays,
		&leave.Reason, &leave.Status, &leave.RequiredApprovals,
		&leave.CreatedAt, &leave.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("查询请假记录失败: %w", err)
	}
	return leave, nil
}

// UpdateStatus 更新请假单状态
func (r *LeaveRepository) UpdateStatus(ctx context.Context, id int64, status models.LeaveStatus) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE leaves SET status=$1, updated_at=$2 WHERE id=$3`,
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新请假状态失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("请假记录不存在: id=%d", id)
	}
	return nil
}

// List 分页查询请假列表
func (r *LeaveRepository) List(ctx context.Context, f *models.LeaveListFilter) ([]*models.Leave, int64, error) {
	args := []any{}
	cond := "WHERE 1=1"
	i := 1

	if f.UserID > 0 {
		cond += fmt.Sprintf(" AND user_id=$%d", i)
		args = append(args, f.UserID)
		i++
	}
	if f.DeptID > 0 {
		cond += fmt.Sprintf(" AND dept_id=$%d", i)
		args = append(args, f.DeptID)
		i++
	}
	if f.Status != "" {
		cond += fmt.Sprintf(" AND status=$%d", i)
		args = append(args, f.Status)
		i++
	}
	if !f.StartFrom.IsZero() {
		cond += fmt.Sprintf(" AND start_date >= $%d", i)
		args = append(args, f.StartFrom)
		i++
	}
	if !f.StartTo.IsZero() {
		cond += fmt.Sprintf(" AND start_date < $%d", i)
		args = append(args, f.StartTo)
		i++
	}

	var total int64
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM leaves "+cond, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计请假记录数失败: %w", err)
	}

	if f.PageSize <= 0 {
		f.PageSize = 20
	}
	offset := (f.Page - 1) * f.PageSize
	args = append(args, f.PageSize, offset)
	query := fmt.Sprintf(`
		SELECT id, user_id, dept_id, leave_type, start_date, end_date,
		       working_days, reason, status, required_approvals, created_at, updated_at
		FROM leaves %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		cond, i, i+1,
	)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询请假列表失败: %w", err)
	}
	defer rows.Close()

	var leaves []*models.Leave
	for rows.Next() {
		l := &models.Leave{}
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.DeptID, &l.LeaveType,
			&l.StartDate, &l.EndDate, &l.WorkingDays,
			&l.Reason, &l.Status, &l.RequiredApprovals,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		leaves = append(leaves, l)
	}
	return leaves, total, nil
}

// CountOnLeaveOnDate 统计某部门在指定日期已批准在假的人数
// 用于检查是否满足部门最低在岗约束
func (r *LeaveRepository) CountOnLeaveOnDate(ctx context.Context, deptID int64, date time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id)
		FROM leaves
		WHERE dept_id = $1
		  AND status = 'approved'
		  AND start_date <= $2
		  AND end_date >= $2`,
		deptID, date,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计在假人数失败: %w", err)
	}
	return count, nil
}

// HasOverlap 检查同一用户在时间段内是否已存在未取消的请假
func (r *LeaveRepository) HasOverlap(ctx context.Context, userID int64, start, end time.Time, excludeID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM leaves
		WHERE user_id = $1
		  AND id != $2
		  AND status NOT IN ('rejected','cancelled')
		  AND start_date <= $4
		  AND end_date >= $3`,
		userID, excludeID, start, end,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查日期冲突失败: %w", err)
	}
	return count > 0, nil
}

// GetPendingApprovedDaysByType 获取某用户指定类型已批准但未来的请假天数（余额预占）
func (r *LeaveRepository) GetPendingApprovedDaysByType(ctx context.Context, userID int64, leaveType models.LeaveType) (float64, error) {
	var days float64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(working_days), 0)
		FROM leaves
		WHERE user_id = $1
		  AND leave_type = $2
		  AND status = 'approved'
		  AND start_date > CURRENT_DATE`,
		userID, leaveType,
	).Scan(&days)
	if err != nil {
		return 0, fmt.Errorf("查询预占余额失败: %w", err)
	}
	return days, nil
}
