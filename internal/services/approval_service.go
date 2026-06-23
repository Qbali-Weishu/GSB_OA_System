package services

import (
	"context"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
	"go.uber.org/zap"
)

// ApprovalService 审批操作服务
type ApprovalService struct {
	approvalRepo *repositories.ApprovalRepository
	leaveRepo    *repositories.LeaveRepository
	workflowSvc  *WorkflowService
	logger       *zap.Logger
}

func NewApprovalService(
	approvalRepo *repositories.ApprovalRepository,
	leaveRepo *repositories.LeaveRepository,
	workflowSvc *WorkflowService,
	logger *zap.Logger,
) *ApprovalService {
	return &ApprovalService{
		approvalRepo: approvalRepo,
		leaveRepo:    leaveRepo,
		workflowSvc:  workflowSvc,
		logger:       logger,
	}
}

// Act 审批人对待办审批节点执行通过或拒绝操作
func (s *ApprovalService) Act(ctx context.Context, approvalID, approverID int64, req *models.ApprovalActionRequest) error {
	if err := s.workflowSvc.ProcessManualApproval(ctx, approvalID, approverID, req); err != nil {
		return fmt.Errorf("审批操作失败: %w", err)
	}
	s.logger.Info("审批操作完成",
		zap.Int64("approval_id", approvalID),
		zap.Int64("approver_id", approverID),
		zap.String("decision", req.Decision),
	)
	return nil
}

// GetPendingList 查询指定审批人的待处理审批列表
func (s *ApprovalService) GetPendingList(ctx context.Context, approverID int64) ([]*models.Approval, error) {
	approvals, err := s.approvalRepo.GetPendingByApprover(ctx, approverID)
	if err != nil {
		return nil, fmt.Errorf("查询待审批列表失败: %w", err)
	}
	return approvals, nil
}

// GetLeaveApprovals 查询请假单的所有审批节点
func (s *ApprovalService) GetLeaveApprovals(ctx context.Context, leaveID int64) ([]*models.Approval, error) {
	approvals, err := s.approvalRepo.GetByLeaveID(ctx, leaveID)
	if err != nil {
		return nil, fmt.Errorf("查询审批记录失败: %w", err)
	}
	return approvals, nil
}

// NotificationRepository 通知仓库接口
type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	UpdateStatus(ctx context.Context, id int64, status models.NotificationStatus) error
	GetPendingBatch(ctx context.Context, limit int) ([]*models.Notification, error)
}

// NotificationService 通知创建服务
type NotificationService struct {
	repo   NotificationRepository
	logger *zap.Logger
}

func NewNotificationService(repo NotificationRepository, logger *zap.Logger) *NotificationService {
	return &NotificationService{repo: repo, logger: logger}
}

// NotifyLeaveApproved 请假批准后通知申请人
func (s *NotificationService) NotifyLeaveApproved(ctx context.Context, leave *models.Leave) {
	now := time.Now()
	n := &models.Notification{
		ReceiverID: leave.UserID,
		RefID:      leave.ID,
		RefType:    "leave",
		Type:       models.NotificationTypeLeaveApproved,
		Title:      "请假申请已批准",
		Content:    fmt.Sprintf("您于 %s 至 %s 的请假申请已审批通过", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")),
		Status:     models.NotificationStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		s.logger.Warn("创建审批通知失败", zap.Int64("leave_id", leave.ID), zap.Error(err))
	}
}

// NotifyLeaveRejected 请假被拒绝后通知申请人
func (s *NotificationService) NotifyLeaveRejected(ctx context.Context, leave *models.Leave, reason string) {
	now := time.Now()
	n := &models.Notification{
		ReceiverID: leave.UserID,
		RefID:      leave.ID,
		RefType:    "leave",
		Type:       models.NotificationTypeLeaveRejected,
		Title:      "请假申请被拒绝",
		Content:    fmt.Sprintf("您的请假申请已被拒绝，原因：%s", reason),
		Status:     models.NotificationStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		s.logger.Warn("创建拒绝通知失败", zap.Int64("leave_id", leave.ID), zap.Error(err))
	}
}

// NotifyApprovalRequest 通知审批人有待审批任务
func (s *NotificationService) NotifyApprovalRequest(ctx context.Context, leave *models.Leave, step *models.Approval) {
	now := time.Now()
	n := &models.Notification{
		ReceiverID: step.ApproverID,
		RefID:      leave.ID,
		RefType:    "leave",
		Type:       models.NotificationTypeApprovalRequest,
		Title:      "您有一条待审批的请假申请",
		Content:    fmt.Sprintf("员工 ID %d 提交了 %s 至 %s 的请假申请，请及时处理", leave.UserID, leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")),
		Status:     models.NotificationStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		s.logger.Warn("创建待办通知失败", zap.Int64("leave_id", leave.ID), zap.Error(err))
	}
}
