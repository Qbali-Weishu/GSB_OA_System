package models

import "time"

// Department 部门
type Department struct {
	ID               int64     `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	ParentID         *int64    `json:"parent_id,omitempty" db:"parent_id"`
	ManagerID        *int64    `json:"manager_id,omitempty" db:"manager_id"`
	HeadCount        int       `json:"head_count" db:"head_count"`
	// MinOnSite 部门最低在岗人数，并发请假时的约束阈值
	MinOnSite        int       `json:"min_on_site" db:"min_on_site"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// DepartmentTree 带层级结构的部门信息
type DepartmentTree struct {
	Department
	Children []*DepartmentTree `json:"children,omitempty"`
}
