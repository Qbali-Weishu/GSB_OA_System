package models

import "time"

// LeaveType 请假类型
type LeaveType string

const (
	LeaveTypeAnnual      LeaveType = "annual"       // 年假
	LeaveTypeSick        LeaveType = "sick"          // 病假
	LeaveTypeCompassion  LeaveType = "compassion"    // 事假
	LeaveTypeCompensatory LeaveType = "compensatory" // 调休
	LeaveTypeMaternity   LeaveType = "maternity"     // 产假
)

// LeaveStatus 请假单状态
type LeaveStatus string

const (
	LeaveStatusPending   LeaveStatus = "pending"   // 待审批
	LeaveStatusApproved  LeaveStatus = "approved"  // 已批准
	LeaveStatusRejected  LeaveStatus = "rejected"  // 已拒绝
	LeaveStatusCancelled LeaveStatus = "cancelled" // 已撤销
	LeaveStatusExpired   LeaveStatus = "expired"   // 已过期
)

// Leave 请假申请单
type Leave struct {
	ID               int64       `json:"id" db:"id"`
	UserID           int64       `json:"user_id" db:"user_id"`
	DeptID           int64       `json:"dept_id" db:"dept_id"`
	LeaveType        LeaveType   `json:"leave_type" db:"leave_type"`
	StartDate        time.Time   `json:"start_date" db:"start_date"`
	EndDate          time.Time   `json:"end_date" db:"end_date"`
	WorkingDays      int         `json:"working_days" db:"working_days"`
	Reason           string      `json:"reason" db:"reason"`
	Status           LeaveStatus `json:"status" db:"status"`
	RequiredApprovals int        `json:"required_approvals" db:"required_approvals"`
	CreatedAt        time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at" db:"updated_at"`
}

// CreateLeaveRequest 提交请假申请的请求体
type CreateLeaveRequest struct {
	LeaveType string `json:"leave_type" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	Reason    string `json:"reason" binding:"required,min=5,max=500"`
}

// LeaveListFilter 请假列表查询条件
type LeaveListFilter struct {
	UserID    int64
	DeptID    int64
	Status    LeaveStatus
	StartFrom time.Time
	StartTo   time.Time
	Page      int
	PageSize  int
}
