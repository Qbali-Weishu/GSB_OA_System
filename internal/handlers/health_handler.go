package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler 健康检查接口
type HealthHandler struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewHealthHandler(db *pgxpool.Pool, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

// Check 检查服务及依赖组件的健康状态
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"checks":    gin.H{},
	}

	checks := status["checks"].(gin.H)
	overall := true

	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = gin.H{"status": "fail", "error": err.Error()}
		overall = false
	} else {
		checks["database"] = gin.H{"status": "ok"}
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		checks["redis"] = gin.H{"status": "fail", "error": err.Error()}
		overall = false
	} else {
		checks["redis"] = gin.H{"status": "ok"}
	}

	if !overall {
		status["status"] = "degraded"
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}
	c.JSON(http.StatusOK, status)
}
