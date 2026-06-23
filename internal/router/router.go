package router

import (
	"time"

	"github.com/company/oa-leave-system/internal/cache"
	"github.com/company/oa-leave-system/internal/config"
	"github.com/company/oa-leave-system/internal/handlers"
	"github.com/company/oa-leave-system/internal/middleware"
	"github.com/company/oa-leave-system/internal/repositories"
	"github.com/company/oa-leave-system/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// New 初始化所有依赖并注册路由
func New(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化仓库
	userRepo := repositories.NewUserRepository(db)
	deptRepo := repositories.NewDepartmentRepository(db)
	leaveRepo := repositories.NewLeaveRepository(db)
	approvalRepo := repositories.NewApprovalRepository(db)
	holidayRepo := repositories.NewHolidayRepository(db)
	balanceRepo := repositories.NewLeaveBalanceRepository(db)
	notifRepo := repositories.NewNotificationRepository(db)

	// 初始化缓存
	holidayCache := cache.NewHolidayCache(24 * time.Hour)
	userCache := cache.NewUserCache(5 * time.Minute)

	// 初始化服务
	holidaySvc := services.NewHolidayService(holidayRepo, holidayCache)
	durationCalc := services.NewDurationCalculator(holidaySvc)
	deptSvc := services.NewDepartmentService(deptRepo, leaveRepo, &cfg.Workflow)
	notifySvc := services.NewNotificationService(notifRepo, logger)
	workflowSvc := services.NewWorkflowService(
		leaveRepo, approvalRepo, deptSvc, notifySvc, logger, cfg.Workflow.AutoApproveMaxDays,
	)
	balanceSvc := services.NewBalanceService(balanceRepo, leaveRepo)
	leaveSvc := services.NewLeaveService(
		leaveRepo, userRepo, deptRepo, balanceSvc, workflowSvc, durationCalc, &cfg.Workflow,
	)
	approvalSvc := services.NewApprovalService(approvalRepo, leaveRepo, workflowSvc, logger)

	// 初始化 Handler
	leaveH := handlers.NewLeaveHandler(leaveSvc)
	approvalH := handlers.NewApprovalHandler(approvalSvc)
	userH := handlers.NewUserHandler(userRepo, cfg.JWT)
	healthH := handlers.NewHealthHandler(db, rdb)

	// 认证中间件
	authMW := middleware.Auth(cfg.JWT.Secret, userRepo, userCache, logger)

	r := gin.New()
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))

	// 公开接口
	r.GET("/health", healthH.Check)
	r.POST("/api/v1/auth/login", userH.Login)

	// 需要认证的接口
	api := r.Group("/api/v1", authMW)
	{
		api.GET("/users/me", userH.Profile)

		// 请假相关
		leaves := api.Group("/leaves")
		{
			leaves.POST("", leaveH.Create)
			leaves.GET("", leaveH.List)
			leaves.GET("/:id", leaveH.GetByID)
			leaves.DELETE("/:id", leaveH.Cancel)
			leaves.GET("/:leave_id/approvals", approvalH.GetLeaveApprovals)
		}

		// 审批相关
		approvals := api.Group("/approvals")
		{
			approvals.GET("/pending", approvalH.PendingList)
			approvals.POST("/:id/act", approvalH.Act)
		}
	}

	return r
}
