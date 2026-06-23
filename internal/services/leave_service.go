package services

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/config"
	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
)

// DurationCalculator 工作日计算服务
type DurationCalculator struct {
	holidaySvc *HolidayService
}

func NewDurationCalculator(holidaySvc *HolidayService) *DurationCalculator {
	return &DurationCalculator{holidaySvc: holidaySvc}
}

// CalculateWorkingDays 计算两个日期之间的工作日天数（含首尾两端）
// 工作日定义：非周末 且 非法定节假日；调班补班日算工作日
func (c *DurationCalculator) CalculateWorkingDays(ctx context.Context, start, end time.Time) (int, error) {
	start = truncateToDay(start)
	end = truncateToDay(end)

	if end.Before(start) {
		return 0, fmt.Errorf("结束日期 %s 早于开始日期 %s", formatDay(end), formatDay(start))
	}

	// 按年份分批加载节假日数据，支持跨年请假
	holidaysByYear := make(map[int]map[string]bool)
	for yr := start.Year(); yr <= end.Year(); yr++ {
		m, err := c.holidaySvc.GetHolidaysByYear(ctx, yr)
		if err != nil {
			return 0, fmt.Errorf("获取 %d 年节假日数据失败: %w", yr, err)
		}
		holidaysByYear[yr] = m
	}

	count := 0
	for cur := start; !cur.After(end); cur = cur.AddDate(0, 0, 1) {
		holidays := holidaysByYear[cur.Year()]
		key := cur.Format("2006-01-02")
		wd := cur.Weekday()
		isWeekend := wd == time.Saturday || wd == time.Sunday

		if isWeekend {
			// 周末仅补班日计为工作日：key 存在且 value=false 代表调班补班
			if val, exists := holidays[key]; exists && !val {
				count++
			}
		} else {
			// 平日：不在节假日集合（value=true）中则计为工作日
			if !holidays[key] {
				count++
			}
		}
	}
	return count, nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func formatDay(t time.Time) string {
	return t.Format("2006-01-02")
}

// LeaveService 请假业务服务
type LeaveService struct {
	leaveRepo    *repositories.LeaveRepository
	userRepo     *repositories.UserRepository
	deptRepo     *repositories.DepartmentRepository
	balanceSvc   *BalanceService
	workflowSvc  *WorkflowService
	durationCalc *DurationCalculator
	cfg          *config.WorkflowConfig
}

func NewLeaveService(
	leaveRepo *repositories.LeaveRepository,
	userRepo *repositories.UserRepository,
	deptRepo *repositories.DepartmentRepository,
	balanceSvc *BalanceService,
	workflowSvc *WorkflowService,
	durationCalc *DurationCalculator,
	cfg *config.WorkflowConfig,
) *LeaveService {
	return &LeaveService{
		leaveRepo:    leaveRepo,
		userRepo:     userRepo,
		deptRepo:     deptRepo,
		balanceSvc:   balanceSvc,
		workflowSvc:  workflowSvc,
		durationCalc: durationCalc,
		cfg:          cfg,
	}
}

// Create 提交请假申请
func (s *LeaveService) Create(ctx context.Context, userID int64, req *models.CreateLeaveRequest) (*models.Leave, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("开始日期格式无效: %w", err)
	}
	endDate, err := time.ParseInLocation("2006-01-02", req.EndDate, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("结束日期格式无效: %w", err)
	}

	if endDate.Before(startDate) {
		return nil, fmt.Errorf("结束日期不能早于开始日期")
	}

	// 检查是否与已有请假存在日期冲突
	overlap, err := s.leaveRepo.HasOverlap(ctx, userID, startDate, endDate, 0)
	if err != nil {
		return nil, fmt.Errorf("检查日期冲突失败: %w", err)
	}
	if overlap {
		return nil, fmt.Errorf("所选日期与已有请假存在冲突，请检查后重新提交")
	}

	// 计算工作日天数
	workingDays, err := s.durationCalc.CalculateWorkingDays(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("计算工作日失败: %w", err)
	}
	if workingDays == 0 {
		return nil, fmt.Errorf("所选日期范围内没有工作日，无需提交请假")
	}

	leaveType := models.LeaveType(req.LeaveType)

	// 校验余额是否充足
	sufficient, err := s.balanceSvc.HasSufficientBalance(ctx, userID, leaveType, float64(workingDays))
	if err != nil {
		return nil, fmt.Errorf("检查余额失败: %w", err)
	}
	if !sufficient {
		return nil, fmt.Errorf("假期余额不足，当前申请 %d 个工作日", workingDays)
	}

	// 根据请假时长确定需要多少级人工审批（自动审批时为 0）
	var requiredApprovals int
	if workingDays > s.cfg.AutoApproveMaxDays {
		requiredApprovals = calculateRequiredApprovals(workingDays)
	}

	now := time.Now()
	leave := &models.Leave{
		UserID:            userID,
		DeptID:            user.DeptID,
		LeaveType:         leaveType,
		StartDate:         startDate,
		EndDate:           endDate,
		WorkingDays:       workingDays,
		Reason:            req.Reason,
		Status:            models.LeaveStatusPending,
		RequiredApprovals: requiredApprovals,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.leaveRepo.Create(ctx, leave); err != nil {
		return nil, fmt.Errorf("保存请假记录失败: %w", err)
	}

	// 触发审批流程（自动或人工）
	if err := s.workflowSvc.Start(ctx, leave); err != nil {
		return nil, fmt.Errorf("启动审批流程失败: %w", err)
	}

	return leave, nil
}

// calculateRequiredApprovals 根据请假天数确定需要的人工审批级别数
func calculateRequiredApprovals(days int) int {
	switch {
	case days <= 3:
		return 1
	case days <= 7:
		return 2
	default:
		return 3
	}
}

// GetByID 查询请假单详情，校验访问权限
func (s *LeaveService) GetByID(ctx context.Context, leaveID, requesterID int64, requesterRole models.Role) (*models.Leave, error) {
	leave, err := s.leaveRepo.GetByID(ctx, leaveID)
	if err != nil {
		return nil, fmt.Errorf("请假记录不存在: %w", err)
	}
	if requesterRole == models.RoleEmployee && leave.UserID != requesterID {
		return nil, fmt.Errorf("无权访问该请假记录")
	}
	return leave, nil
}

// Cancel 撤销请假申请（仅限申请人在审批通过前操作）
func (s *LeaveService) Cancel(ctx context.Context, leaveID, userID int64) error {
	leave, err := s.leaveRepo.GetByID(ctx, leaveID)
	if err != nil {
		return fmt.Errorf("请假记录不存在: %w", err)
	}
	if leave.UserID != userID {
		return fmt.Errorf("无权撤销他人的请假申请")
	}
	if leave.Status != models.LeaveStatusPending {
		return fmt.Errorf("仅待审批状态的请假单可以撤销，当前状态: %s", leave.Status)
	}
	return s.leaveRepo.UpdateStatus(ctx, leaveID, models.LeaveStatusCancelled)
}

// List 分页查询请假列表
func (s *LeaveService) List(ctx context.Context, filter *models.LeaveListFilter) ([]*models.Leave, int64, error) {
	return s.leaveRepo.List(ctx, filter)
}
