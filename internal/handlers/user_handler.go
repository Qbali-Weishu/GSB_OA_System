package handlers

import (
	"net/http"
	"time"

	"github.com/company/oa-leave-system/internal/middleware"
	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/repositories"
	"github.com/company/oa-leave-system/internal/utils"
	"github.com/company/oa-leave-system/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler 用户相关 HTTP 接口
type UserHandler struct {
	userRepo *repositories.UserRepository
	jwtCfg   config.JWTConfig
}

func NewUserHandler(userRepo *repositories.UserRepository, jwtCfg config.JWTConfig) *UserHandler {
	return &UserHandler{userRepo: userRepo, jwtCfg: jwtCfg}
}

// Login 用户登录，返回 JWT 令牌
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailMsg(c, http.StatusBadRequest, "请求参数无效")
		return
	}
	user, err := h.userRepo.GetByUsername(c.Request.Context(), req.Username)
	if err != nil {
		utils.FailMsg(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	if !user.IsActive {
		utils.FailMsg(c, http.StatusForbidden, "账号已被停用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		utils.FailMsg(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	expiresAt := time.Now().Add(h.jwtCfg.Expiration)
	claims := &middleware.Claims{
		UserID: user.ID,
		Role:   user.Role,
		DeptID: user.DeptID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(h.jwtCfg.Secret))
	if err != nil {
		utils.FailMsg(c, http.StatusInternalServerError, "生成令牌失败")
		return
	}

	utils.OK(c, models.TokenResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
		User: models.UserProfile{
			ID:     user.ID,
			Name:   user.Name,
			Email:  user.Email,
			DeptID: user.DeptID,
			Role:   user.Role,
		},
	})
}

// Profile 查询当前登录用户的个人信息
func (h *UserHandler) Profile(c *gin.Context) {
	user := middleware.CurrentUser(c)
	profile, err := h.userRepo.GetProfile(c.Request.Context(), user.ID)
	if err != nil {
		utils.FailMsg(c, http.StatusInternalServerError, "查询用户信息失败")
		return
	}
	utils.OK(c, profile)
}
