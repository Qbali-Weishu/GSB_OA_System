package middleware

import (
	"net/http"
	"strings"

	"github.com/company/oa-leave-system/internal/cache"
	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

const ContextKeyUser = "authenticated_user"

// Claims JWT载荷
type Claims struct {
	UserID int64       `json:"user_id"`
	Role   models.Role `json:"role"`
	DeptID int64       `json:"dept_id"`
	jwt.RegisteredClaims
}

// Auth 认证中间件，校验 Bearer Token 并将用户信息注入上下文
func Auth(secret string, userRepo *repositories.UserRepository, userCache *cache.UserCache, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少认证令牌"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "令牌无效或已过期"})
			return
		}

		// 优先从缓存读取用户信息，减少数据库查询
		user, ok := userCache.Get(claims.UserID)
		if !ok {
			user, err = userRepo.GetByID(c.Request.Context(), claims.UserID)
			if err != nil {
				logger.Warn("认证用户不存在", zap.Int64("user_id", claims.UserID), zap.Error(err))
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "用户不存在或已被停用"})
				return
			}
			userCache.Set(user)
		}

		c.Set(ContextKeyUser, user)
		c.Next()
	}
}

// RequireRole 角色权限中间件，限制接口仅特定角色可访问
func RequireRole(roles ...models.Role) gin.HandlerFunc {
	allowed := make(map[models.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		user, exists := c.Get(ContextKeyUser)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
			return
		}
		u := user.(*models.User)
		if !allowed[u.Role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足"})
			return
		}
		c.Next()
	}
}

// CurrentUser 从 gin 上下文中取出当前用户，调用前确保已通过 Auth 中间件
func CurrentUser(c *gin.Context) *models.User {
	u, _ := c.Get(ContextKeyUser)
	return u.(*models.User)
}
