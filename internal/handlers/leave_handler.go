package handlers

import (
	"net/http"
	"strconv"

	"github.com/company/oa-leave-system/internal/middleware"
	"github.com/company/oa-leave-system/internal/models"
	"github.com/company/oa-leave-system/internal/services"
	"github.com/company/oa-leave-system/internal/utils"
	"github.com/gin-gonic/gin"
)

// LeaveHandler 请假相关 HTTP 接口
type LeaveHandler struct {
	leaveSvc *services.LeaveService
}

func NewLeaveHandler(leaveSvc *services.LeaveService) *LeaveHandler {
	return &LeaveHandler{leaveSvc: leaveSvc}
}

// Create 提交请假申请
func (h *LeaveHandler) Create(c *gin.Context) {
	user := middleware.CurrentUser(c)
	var req models.CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailMsg(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}
	leave, err := h.leaveSvc.Create(c.Request.Context(), user.ID, &req)
	if err != nil {
		if ae, ok := utils.IsAppError(err); ok {
			utils.Fail(c, ae)
			return
		}
		utils.FailMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.OK(c, leave)
}

// GetByID 查询单条请假记录
func (h *LeaveHandler) GetByID(c *gin.Context) {
	user := middleware.CurrentUser(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.FailMsg(c, http.StatusBadRequest, "无效的请假 ID")
		return
	}
	leave, err := h.leaveSvc.GetByID(c.Request.Context(), id, user.ID, user.Role)
	if err != nil {
		utils.FailMsg(c, http.StatusNotFound, err.Error())
		return
	}
	utils.OK(c, leave)
}

// List 分页查询请假列表
func (h *LeaveHandler) List(c *gin.Context) {
	user := middleware.CurrentUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := &models.LeaveListFilter{
		Page:     page,
		PageSize: pageSize,
	}
	// 普通员工只能查自己的记录，HR 和管理员可按部门查询
	if user.Role == models.RoleEmployee {
		filter.UserID = user.ID
	} else if deptStr := c.Query("dept_id"); deptStr != "" {
		filter.DeptID, _ = strconv.ParseInt(deptStr, 10, 64)
	}
	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = models.LeaveStatus(statusStr)
	}

	leaves, total, err := h.leaveSvc.List(c.Request.Context(), filter)
	if err != nil {
		utils.FailMsg(c, http.StatusInternalServerError, "查询失败")
		return
	}
	utils.OKPaged(c, leaves, total, page, pageSize)
}

// Cancel 撤销请假申请
func (h *LeaveHandler) Cancel(c *gin.Context) {
	user := middleware.CurrentUser(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.FailMsg(c, http.StatusBadRequest, "无效的请假 ID")
		return
	}
	if err := h.leaveSvc.Cancel(c.Request.Context(), id, user.ID); err != nil {
		utils.FailMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.OK(c, gin.H{"message": "已撤销"})
}
