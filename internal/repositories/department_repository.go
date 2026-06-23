package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DepartmentRepository 部门数据访问层
type DepartmentRepository struct {
	db *pgxpool.Pool
}

func NewDepartmentRepository(db *pgxpool.Pool) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// GetByID 查询部门详情
func (r *DepartmentRepository) GetByID(ctx context.Context, id int64) (*models.Department, error) {
	dept := &models.Department{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, parent_id, manager_id, head_count, min_on_site, created_at, updated_at
		FROM departments WHERE id = $1`, id,
	).Scan(
		&dept.ID, &dept.Name, &dept.ParentID, &dept.ManagerID,
		&dept.HeadCount, &dept.MinOnSite, &dept.CreatedAt, &dept.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("查询部门失败: %w", err)
	}
	return dept, nil
}

// UpdateHeadCount 同步更新部门人数（员工入离职时调用）
func (r *DepartmentRepository) UpdateHeadCount(ctx context.Context, id int64, delta int) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE departments
		SET head_count = head_count + $1, updated_at = $2
		WHERE id = $3`,
		delta, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新部门人数失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("部门不存在: id=%d", id)
	}
	return nil
}

// ListAll 查询所有部门列表（构建部门树用）
func (r *DepartmentRepository) ListAll(ctx context.Context) ([]*models.Department, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, parent_id, manager_id, head_count, min_on_site, created_at, updated_at
		FROM departments ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("查询部门列表失败: %w", err)
	}
	defer rows.Close()

	var depts []*models.Department
	for rows.Next() {
		d := &models.Department{}
		if err := rows.Scan(
			&d.ID, &d.Name, &d.ParentID, &d.ManagerID,
			&d.HeadCount, &d.MinOnSite, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		depts = append(depts, d)
	}
	return depts, nil
}
