package services

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/config"
	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
)

// DepartmentService 部门业务逻辑服务
type DepartmentService struct {
	deptRepo  *repositories.DepartmentRepository
	leaveRepo *repositories.LeaveRepository
	cfg       *config.WorkflowConfig
}

func NewDepartmentService(
	deptRepo *repositories.DepartmentRepository,
	leaveRepo *repositories.LeaveRepository,
	cfg *config.WorkflowConfig,
) *DepartmentService {
	return &DepartmentService{deptRepo: deptRepo, leaveRepo: leaveRepo, cfg: cfg}
}

// CheckMinimumOnSite 检查部门在指定日期是否满足最低在岗人数约束
// 返回 true 表示当前在岗人数充裕，批准一人休假后仍能满足约束
func (s *DepartmentService) CheckMinimumOnSite(ctx context.Context, deptID int64, leaveDate time.Time) (bool, error) {
	dept, err := s.deptRepo.GetByID(ctx, deptID)
	if err != nil {
		return false, fmt.Errorf("获取部门信息失败: %w", err)
	}

	// 统计该日期已批准的在假人数（start_date <= date <= end_date 且 status=approved）
	onLeaveCount, err := s.leaveRepo.CountOnLeaveOnDate(ctx, deptID, leaveDate)
	if err != nil {
		return false, fmt.Errorf("统计在假人数失败: %w", err)
	}

	// 取部门自定义阈值与全局配置的较大值作为最低在岗要求
	minRequired := dept.MinOnSite
	if minRequired < s.cfg.MinimumOnSiteCount {
		minRequired = s.cfg.MinimumOnSiteCount
	}

	// 当前在岗人数须严格大于最低阈值，确保批假后在岗人数仍不低于阈值
	currentOnSite := dept.HeadCount - onLeaveCount
	return currentOnSite > minRequired, nil
}

// GetTree 返回部门层级树
func (s *DepartmentService) GetTree(ctx context.Context) ([]*models.DepartmentTree, error) {
	all, err := s.deptRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询部门列表失败: %w", err)
	}
	return buildTree(all), nil
}

func buildTree(depts []*models.Department) []*models.DepartmentTree {
	nodeMap := make(map[int64]*models.DepartmentTree)
	for _, d := range depts {
		nodeMap[d.ID] = &models.DepartmentTree{Department: *d}
	}
	var roots []*models.DepartmentTree
	for _, node := range nodeMap {
		if node.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}
	return roots
}
