package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationRepository 通知数据访问层
type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create 写入一条待发送通知
func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO notifications
		  (receiver_id, sender_id, ref_id, ref_type, type, title, content, status, retry_count, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id`,
		n.ReceiverID, n.SenderID, n.RefID, n.RefType, n.Type,
		n.Title, n.Content, n.Status, n.RetryCount, n.CreatedAt, n.UpdatedAt,
	).Scan(&n.ID)
	if err != nil {
		return fmt.Errorf("创建通知记录失败: %w", err)
	}
	return nil
}

// UpdateStatus 更新通知发送状态
func (r *NotificationRepository) UpdateStatus(ctx context.Context, id int64, status models.NotificationStatus) error {
	now := time.Now()
	var sentAt *time.Time
	if status == models.NotificationStatusSent {
		sentAt = &now
	}
	_, err := r.db.Exec(ctx, `
		UPDATE notifications SET status=$1, sent_at=$2, updated_at=$3 WHERE id=$4`,
		status, sentAt, now, id,
	)
	if err != nil {
		return fmt.Errorf("更新通知状态失败: %w", err)
	}
	return nil
}

// GetPendingBatch 获取一批待发送通知（按创建时间升序，最多 limit 条）
func (r *NotificationRepository) GetPendingBatch(ctx context.Context, limit int) ([]*models.Notification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, receiver_id, sender_id, ref_id, ref_type, type, title, content,
		       status, retry_count, sent_at, created_at, updated_at
		FROM notifications
		WHERE status = 'pending' AND retry_count < 3
		ORDER BY created_at ASC
		LIMIT $1`, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("查询待发通知失败: %w", err)
	}
	defer rows.Close()

	var list []*models.Notification
	for rows.Next() {
		n := &models.Notification{}
		if err := rows.Scan(
			&n.ID, &n.ReceiverID, &n.SenderID, &n.RefID, &n.RefType,
			&n.Type, &n.Title, &n.Content, &n.Status, &n.RetryCount,
			&n.SentAt, &n.CreatedAt, &n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, nil
}

// IncrRetryCount 增加重试计数（发送失败时调用）
func (r *NotificationRepository) IncrRetryCount(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE notifications SET retry_count = retry_count + 1, updated_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}
