package models

import "time"

// Role 用户角色
type Role string

const (
	RoleEmployee  Role = "employee"  // 普通员工
	RoleManager   Role = "manager"   // 部门经理
	RoleHR        Role = "hr"        // HR
	RoleAdmin     Role = "admin"     // 系统管理员
)

// User 系统用户
type User struct {
	ID         int64     `json:"id" db:"id"`
	Username   string    `json:"username" db:"username"`
	PasswordHash string  `json:"-" db:"password_hash"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Phone      string    `json:"phone" db:"phone"`
	DeptID     int64     `json:"dept_id" db:"dept_id"`
	ManagerID  *int64    `json:"manager_id,omitempty" db:"manager_id"`
	Role       Role      `json:"role" db:"role"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	JoinDate   time.Time `json:"join_date" db:"join_date"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile 对外暴露的用户摘要
type UserProfile struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	DeptID   int64  `json:"dept_id"`
	DeptName string `json:"dept_name"`
	Role     Role   `json:"role"`
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse 登录成功返回的令牌信息
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserProfile `json:"user"`
}
