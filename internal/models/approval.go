package models

import "time"

// ApprovalStatus 审批节点状态
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"  // 待审批
	ApprovalStatusApproved ApprovalStatus = "approved" // 已批准
	ApprovalStatusRejected ApprovalStatus = "rejected" // 已拒绝
)

// ApproverType 审批人角色类型
type ApproverType string

const (
	ApproverTypeManager ApproverType = "manager" // 直属上级
	ApproverTypeDeptHead ApproverType = "dept_head" // 部门负责人
	ApproverTypeHR      ApproverType = "hr"      // HR
	ApproverTypeSystem  ApproverType = "system"  // 系统自动
)

// Approval 审批节点记录
type Approval struct {
	ID           int64          `json:"id" db:"id"`
	LeaveID      int64          `json:"leave_id" db:"leave_id"`
	ApproverID   int64          `json:"approver_id" db:"approver_id"`
	ApproverType ApproverType   `json:"approver_type" db:"approver_type"`
	StepOrder    int            `json:"step_order" db:"step_order"`
	Status       ApprovalStatus `json:"status" db:"status"`
	Comment      string         `json:"comment" db:"comment"`
	ActedAt      *time.Time     `json:"acted_at,omitempty" db:"acted_at"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

// ApprovalActionRequest 审批操作请求体
type ApprovalActionRequest struct {
	Decision string `json:"decision" binding:"required,oneof=approved rejected"`
	Comment  string `json:"comment" binding:"max=500"`
}
