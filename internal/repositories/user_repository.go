package repositories

import (
	"context"
	"fmt"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID 按主键查询用户，联查部门名称
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.username, u.password_hash, u.name, u.email, u.phone,
		       u.dept_id, u.manager_id, u.role, u.is_active, u.join_date,
		       u.created_at, u.updated_at
		FROM users u
		WHERE u.id = $1 AND u.is_active = true`, id,
	).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Name, &u.Email, &u.Phone,
		&u.DeptID, &u.ManagerID, &u.Role, &u.IsActive, &u.JoinDate,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// GetByUsername 按用户名查询用户（含密码哈希，仅用于登录）
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, username, password_hash, name, email, phone,
		       dept_id, manager_id, role, is_active, join_date, created_at, updated_at
		FROM users
		WHERE username = $1`, username,
	).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Name, &u.Email, &u.Phone,
		&u.DeptID, &u.ManagerID, &u.Role, &u.IsActive, &u.JoinDate,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return u, nil
}

// ListByDept 查询指定部门的在职用户列表
func (r *UserRepository) ListByDept(ctx context.Context, deptID int64) ([]*models.User, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, username, password_hash, name, email, phone,
		       dept_id, manager_id, role, is_active, join_date, created_at, updated_at
		FROM users
		WHERE dept_id = $1 AND is_active = true
		ORDER BY name`, deptID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询部门用户失败: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(
			&u.ID, &u.Username, &u.PasswordHash, &u.Name, &u.Email, &u.Phone,
			&u.DeptID, &u.ManagerID, &u.Role, &u.IsActive, &u.JoinDate,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetProfile 查询用户公开信息（含部门名称）
func (r *UserRepository) GetProfile(ctx context.Context, id int64) (*models.UserProfile, error) {
	p := &models.UserProfile{}
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.dept_id, d.name, u.role
		FROM users u
		JOIN departments d ON d.id = u.dept_id
		WHERE u.id = $1 AND u.is_active = true`, id,
	).Scan(&p.ID, &p.Name, &p.Email, &p.DeptID, &p.DeptName, &p.Role)
	if err != nil {
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}
	return p, nil
}
