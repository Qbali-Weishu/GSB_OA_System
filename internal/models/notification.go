package models

import "time"

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeLeaveSubmitted  NotificationType = "leave_submitted"  // 提交请假通知审批人
	NotificationTypeLeaveApproved   NotificationType = "leave_approved"   // 请假已批准通知申请人
	NotificationTypeLeaveRejected   NotificationType = "leave_rejected"   // 请假被拒绝通知申请人
	NotificationTypeApprovalRequest NotificationType = "approval_request" // 待办审批提醒
)

// NotificationStatus 通知发送状态
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
)

// Notification 待发送或已发送的通知记录
type Notification struct {
	ID         int64              `json:"id" db:"id"`
	ReceiverID int64              `json:"receiver_id" db:"receiver_id"`
	SenderID   *int64             `json:"sender_id,omitempty" db:"sender_id"`
	RefID      int64              `json:"ref_id" db:"ref_id"`     // 关联业务对象 ID（如 leave_id）
	RefType    string             `json:"ref_type" db:"ref_type"` // 关联业务对象类型
	Type       NotificationType   `json:"type" db:"type"`
	Title      string             `json:"title" db:"title"`
	Content    string             `json:"content" db:"content"`
	Status     NotificationStatus `json:"status" db:"status"`
	RetryCount int                `json:"retry_count" db:"retry_count"`
	SentAt     *time.Time         `json:"sent_at,omitempty" db:"sent_at"`
	CreatedAt  time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" db:"updated_at"`
}

// LeaveBalance 用户请假余额（按类型）
type LeaveBalance struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	Year         int       `json:"year" db:"year"`
	LeaveType    LeaveType `json:"leave_type" db:"leave_type"`
	TotalDays    float64   `json:"total_days" db:"total_days"`
	UsedDays     float64   `json:"used_days" db:"used_days"`
	PendingDays  float64   `json:"pending_days" db:"pending_days"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
