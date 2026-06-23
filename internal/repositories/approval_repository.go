package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApprovalRepository 审批节点数据访问层
type ApprovalRepository struct {
	db *pgxpool.Pool
}

func NewApprovalRepository(db *pgxpool.Pool) *ApprovalRepository {
	return &ApprovalRepository{db: db}
}

// Create 创建审批节点记录
func (r *ApprovalRepository) Create(ctx context.Context, a *models.Approval) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO approvals (leave_id, approver_id, approver_type, step_order, status, comment, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`,
		a.LeaveID, a.ApproverID, a.ApproverType, a.StepOrder,
		a.Status, a.Comment, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID)
	if err != nil {
		return fmt.Errorf("创建审批记录失败: %w", err)
	}
	return nil
}

// UpdateStatus 更新审批节点的审批结论
func (r *ApprovalRepository) UpdateStatus(ctx context.Context, id int64, status models.ApprovalStatus, comment string) error {
	now := time.Now()
	tag, err := r.db.Exec(ctx, `
		UPDATE approvals
		SET status=$1, comment=$2, acted_at=$3, updated_at=$3
		WHERE id=$4`,
		status, comment, now, id,
	)
	if err != nil {
		return fmt.Errorf("更新审批状态失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("审批记录不存在: id=%d", id)
	}
	return nil
}

// GetByLeaveID 查询请假单下所有审批节点，按步骤顺序排列
func (r *ApprovalRepository) GetByLeaveID(ctx context.Context, leaveID int64) ([]*models.Approval, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, leave_id, approver_id, approver_type, step_order,
		       status, comment, acted_at, created_at, updated_at
		FROM approvals
		WHERE leave_id = $1
		ORDER BY step_order ASC`, leaveID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询审批记录失败: %w", err)
	}
	defer rows.Close()

	var approvals []*models.Approval
	for rows.Next() {
		a := &models.Approval{}
		if err := rows.Scan(
			&a.ID, &a.LeaveID, &a.ApproverID, &a.ApproverType, &a.StepOrder,
			&a.Status, &a.Comment, &a.ActedAt, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		approvals = append(approvals, a)
	}
	return approvals, nil
}

// GetPendingByApprover 查询指定审批人的待处理审批列表
func (r *ApprovalRepository) GetPendingByApprover(ctx context.Context, approverID int64) ([]*models.Approval, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, a.leave_id, a.approver_id, a.approver_type, a.step_order,
		       a.status, a.comment, a.acted_at, a.created_at, a.updated_at
		FROM approvals a
		JOIN leaves l ON l.id = a.leave_id
		WHERE a.approver_id = $1
		  AND a.status = 'pending'
		  AND l.status = 'pending'
		ORDER BY a.created_at ASC`, approverID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询待审批列表失败: %w", err)
	}
	defer rows.Close()

	var approvals []*models.Approval
	for rows.Next() {
		a := &models.Approval{}
		if err := rows.Scan(
			&a.ID, &a.LeaveID, &a.ApproverID, &a.ApproverType, &a.StepOrder,
			&a.Status, &a.Comment, &a.ActedAt, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		approvals = append(approvals, a)
	}
	return approvals, nil
}

// CountApprovedSteps 统计指定请假单已通过审批的节点数
func (r *ApprovalRepository) CountApprovedSteps(ctx context.Context, leaveID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM approvals WHERE leave_id=$1 AND status='approved'`,
		leaveID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计已审批节点失败: %w", err)
	}
	return count, nil
}
