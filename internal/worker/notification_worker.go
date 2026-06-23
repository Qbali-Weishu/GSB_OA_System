package worker

import (
	"context"
	"time"

	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// NotificationWorker 后台通知发送工作线程
// 每隔 10 秒轮询待发通知，模拟发送（写日志）后更新状态
type NotificationWorker struct {
	notifRepo *repositories.NotificationRepository
	logger    *zap.Logger
	stopCh    chan struct{}
}

func NewNotificationWorker(db *pgxpool.Pool, _ *redis.Client, logger *zap.Logger) *NotificationWorker {
	return &NotificationWorker{
		notifRepo: repositories.NewNotificationRepository(db),
		logger:    logger,
		stopCh:    make(chan struct{}),
	}
}

// Run 启动工作线程主循环
func (w *NotificationWorker) Run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	w.logger.Info("通知工作线程已启动")
	for {
		select {
		case <-ticker.C:
			w.processBatch()
		case <-w.stopCh:
			w.logger.Info("通知工作线程已停止")
			return
		}
	}
}

// Stop 优雅停止工作线程
func (w *NotificationWorker) Stop() {
	close(w.stopCh)
}

func (w *NotificationWorker) processBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	notifications, err := w.notifRepo.GetPendingBatch(ctx, 50)
	if err != nil {
		w.logger.Error("获取待发通知失败", zap.Error(err))
		return
	}

	for _, n := range notifications {
		if err := w.send(ctx, n); err != nil {
			w.logger.Warn("通知发送失败，稍后重试",
				zap.Int64("notification_id", n.ID),
				zap.Int("retry_count", n.RetryCount),
				zap.Error(err),
			)
			w.notifRepo.IncrRetryCount(ctx, n.ID)
		} else {
			w.notifRepo.UpdateStatus(ctx, n.ID, models.NotificationStatusSent)
		}
	}
}

// send 执行实际的通知发送逻辑
// 生产环境中此处会调用邮件/短信/企业微信等通道，当前以日志记录代替
func (w *NotificationWorker) send(ctx context.Context, n *models.Notification) error {
	w.logger.Info("发送通知",
		zap.Int64("id", n.ID),
		zap.Int64("receiver_id", n.ReceiverID),
		zap.String("type", string(n.Type)),
		zap.String("title", n.Title),
	)
	return nil
}
