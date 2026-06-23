package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
	"go.uber.org/zap"
)

const SystemApproverID = int64(0)

var (
	ErrExceedsAutoApproveLimit = errors.New("请假时长超出自动审批上限，已转入人工审批")
	ErrInsufficientOnSite      = errors.New("部门当日在岗人数不足，已转入人工审批")
)

// WorkflowService 审批流程编排服务
type WorkflowService struct {
	leaveRepo    *repositories.LeaveRepository
	approvalRepo *repositories.ApprovalRepository
	deptSvc      *DepartmentService
	notifySvc    *NotificationService
	logger       *zap.Logger
	autoMaxDays  int
}

func NewWorkflowService(
	leaveRepo *repositories.LeaveRepository,
	approvalRepo *repositories.ApprovalRepository,
	deptSvc *DepartmentService,
	notifySvc *NotificationService,
	logger *zap.Logger,
	autoMaxDays int,
) *WorkflowService {
	return &WorkflowService{
		leaveRepo:    leaveRepo,
		approvalRepo: approvalRepo,
		deptSvc:      deptSvc,
		notifySvc:    notifySvc,
		logger:       logger,
		autoMaxDays:  autoMaxDays,
	}
}

// Start 根据请假时长决定走自动审批还是人工审批流程
func (s *WorkflowService) Start(ctx context.Context, leave *models.Leave) error {
	if leave.WorkingDays <= s.autoMaxDays {
		if err := s.AutoApprove(ctx, leave.ID); err != nil {
			if errors.Is(err, ErrInsufficientOnSite) || errors.Is(err, ErrExceedsAutoApproveLimit) {
				// 自动审批条件不满足时平滑降级为人工审批
				s.logger.Info("自动审批条件不满足，转为人工审批",
					zap.Int64("leave_id", leave.ID),
					zap.String("reason", err.Error()),
				)
				return s.initiateManualApproval(ctx, leave)
			}
			return err
		}
		return nil
	}
	return s.initiateManualApproval(ctx, leave)
}

// AutoApprove 对时长符合条件的短假执行系统自动审批
// 自动审批的前提：部门当日在岗人数满足最低约束
func (s *WorkflowService) AutoApprove(ctx context.Context, leaveID int64) error {
	leave, err := s.leaveRepo.GetByID(ctx, leaveID)
	if err != nil {
		return fmt.Errorf("获取请假记录失败: %w", err)
	}

	if leave.Status != models.LeaveStatusPending {
		return fmt.Errorf("请假单状态异常: status=%s", leave.Status)
	}

	if leave.WorkingDays > s.autoMaxDays {
		return ErrExceedsAutoApproveLimit
	}

	// 检查批假后部门在岗人数是否仍满足最低约束
	sufficient, err := s.deptSvc.CheckMinimumOnSite(ctx, leave.DeptID, leave.StartDate)
	if err != nil {
		return fmt.Errorf("检查在岗约束失败: %w", err)
	}
	if !sufficient {
		return ErrInsufficientOnSite
	}

	// 写入系统自动审批记录
	now := time.Now()
	approval := &models.Approval{
		LeaveID:      leaveID,
		ApproverID:   SystemApproverID,
		ApproverType: models.ApproverTypeSystem,
		StepOrder:    1,
		Status:       models.ApprovalStatusApproved,
		Comment:      "系统自动审批通过",
		ActedAt:      &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.approvalRepo.Create(ctx, approval); err != nil {
		return fmt.Errorf("写入审批记录失败: %w", err)
	}

	if err := s.leaveRepo.UpdateStatus(ctx, leaveID, models.LeaveStatusApproved); err != nil {
		return fmt.Errorf("更新请假状态失败: %w", err)
	}

	s.logger.Info("短假自动审批通过",
		zap.Int64("leave_id", leaveID),
		zap.Int64("user_id", leave.UserID),
	)
	go s.notifySvc.NotifyLeaveApproved(context.Background(), leave)
	return nil
}

// initiateManualApproval 发起人工审批流程，按层级创建待审批节点
func (s *WorkflowService) initiateManualApproval(ctx context.Context, leave *models.Leave) error {
	steps := buildApprovalSteps(leave)
	now := time.Now()
	for _, step := range steps {
		step.LeaveID = leave.ID
		step.Status = models.ApprovalStatusPending
		step.CreatedAt = now
		step.UpdatedAt = now
		if err := s.approvalRepo.Create(ctx, step); err != nil {
			return fmt.Errorf("创建审批节点失败 step=%d: %w", step.StepOrder, err)
		}
	}
	go s.notifySvc.NotifyApprovalRequest(context.Background(), leave, steps[0])
	return nil
}

// ProcessManualApproval 处理人工审批人的审批操作
func (s *WorkflowService) ProcessManualApproval(ctx context.Context, approvalID, approverID int64, req *models.ApprovalActionRequest) error {
	approvals, err := s.approvalRepo.GetPendingByApprover(ctx, approverID)
	if err != nil {
		return fmt.Errorf("查询待审批列表失败: %w", err)
	}

	var target *models.Approval
	for _, a := range approvals {
		if a.ID == approvalID {
			target = a
			break
		}
	}
	if target == nil {
		return fmt.Errorf("审批记录不存在或无权操作")
	}

	status := models.ApprovalStatusApproved
	if req.Decision == "rejected" {
		status = models.ApprovalStatusRejected
	}

	if err := s.approvalRepo.UpdateStatus(ctx, approvalID, status, req.Comment); err != nil {
		return fmt.Errorf("更新审批状态失败: %w", err)
	}

	if status == models.ApprovalStatusRejected {
		return s.leaveRepo.UpdateStatus(ctx, target.LeaveID, models.LeaveStatusRejected)
	}

	// 检查是否所有审批节点均已通过
	leave, err := s.leaveRepo.GetByID(ctx, target.LeaveID)
	if err != nil {
		return fmt.Errorf("获取请假记录失败: %w", err)
	}

	approvedCount, err := s.approvalRepo.CountApprovedSteps(ctx, target.LeaveID)
	if err != nil {
		return fmt.Errorf("统计已通过节点失败: %w", err)
	}

	if approvedCount >= leave.RequiredApprovals {
		if err := s.leaveRepo.UpdateStatus(ctx, target.LeaveID, models.LeaveStatusApproved); err != nil {
			return fmt.Errorf("更新请假状态失败: %w", err)
		}
		go s.notifySvc.NotifyLeaveApproved(context.Background(), leave)
	}
	return nil
}

// buildApprovalSteps 根据请假天数构建审批节点列表
func buildApprovalSteps(leave *models.Leave) []*models.Approval {
	steps := []*models.Approval{
		{StepOrder: 1, ApproverType: models.ApproverTypeManager},
	}
	if leave.WorkingDays > 3 {
		steps = append(steps, &models.Approval{StepOrder: 2, ApproverType: models.ApproverTypeDeptHead})
	}
	if leave.WorkingDays > 7 {
		steps = append(steps, &models.Approval{StepOrder: 3, ApproverType: models.ApproverTypeHR})
	}
	return steps
}
