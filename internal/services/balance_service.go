package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
	"github.com/jackc/pgx/v5"
)

// BalanceRepository 余额仓库接口（避免循环依赖）
type BalanceRepository interface {
	GetByUserAndType(ctx context.Context, userID int64, year int, leaveType models.LeaveType) (*models.LeaveBalance, error)
	Deduct(ctx context.Context, userID int64, year int, leaveType models.LeaveType, days float64) error
	Add(ctx context.Context, userID int64, year int, leaveType models.LeaveType, days float64) error
}

// BalanceService 假期余额管理服务
type BalanceService struct {
	balanceRepo BalanceRepository
	leaveRepo   *repositories.LeaveRepository
}

func NewBalanceService(balanceRepo BalanceRepository, leaveRepo *repositories.LeaveRepository) *BalanceService {
	return &BalanceService{balanceRepo: balanceRepo, leaveRepo: leaveRepo}
}

// HasSufficientBalance 检查用户指定类型的可用余额是否满足申请需求
// 可用余额 = 总额 - 已用额 - 已批准尚未开始的假期占用
func (s *BalanceService) HasSufficientBalance(ctx context.Context, userID int64, leaveType models.LeaveType, days float64) (bool, error) {
	// 事假不受余额约束
	if leaveType == models.LeaveTypeCompassion {
		return true, nil
	}

	currentYear := time.Now().Year()
	balance, err := s.balanceRepo.GetByUserAndType(ctx, userID, currentYear, leaveType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("查询余额失败: %w", err)
	}

	// 查询已批准但尚未开始的假期占用额度，防止重复申请超额
	pendingDays, err := s.leaveRepo.GetPendingApprovedDaysByType(ctx, userID, leaveType)
	if err != nil {
		return false, fmt.Errorf("查询占用余额失败: %w", err)
	}

	available := balance.TotalDays - balance.UsedDays - pendingDays
	return available >= days, nil
}

// Deduct 审批通过后正式扣减余额
// 扣减优先级：调休假 → 年假（年假申请时若有足够调休余额则优先消耗调休）
func (s *BalanceService) Deduct(ctx context.Context, userID int64, leaveType models.LeaveType, days float64) error {
	if leaveType == models.LeaveTypeCompassion {
		return nil
	}

	currentYear := time.Now().Year()

	// 年假申请时检查是否有调休余额可优先抵扣
	if leaveType == models.LeaveTypeAnnual {
		compBalance, err := s.balanceRepo.GetByUserAndType(ctx, userID, currentYear, models.LeaveTypeCompensatory)
		if err == nil {
			availableComp := compBalance.TotalDays - compBalance.UsedDays
			if availableComp >= days {
				// 调休余额充足，优先消耗调休额度，年假额度不变
				return s.balanceRepo.Deduct(ctx, userID, currentYear, models.LeaveTypeCompensatory, days)
			}
		}
	}

	return s.balanceRepo.Deduct(ctx, userID, currentYear, leaveType, days)
}

// Refund 撤销或拒绝后退还余额
func (s *BalanceService) Refund(ctx context.Context, userID int64, leaveType models.LeaveType, days float64) error {
	if leaveType == models.LeaveTypeCompassion {
		return nil
	}
	currentYear := time.Now().Year()
	return s.balanceRepo.Add(ctx, userID, currentYear, leaveType, days)
}
