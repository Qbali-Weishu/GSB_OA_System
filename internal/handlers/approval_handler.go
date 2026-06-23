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

// ApprovalHandler 审批相关 HTTP 接口
type ApprovalHandler struct {
	approvalSvc *services.ApprovalService
}

func NewApprovalHandler(approvalSvc *services.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{approvalSvc: approvalSvc}
}

// PendingList 查询当前用户的待审批列表
func (h *ApprovalHandler) PendingList(c *gin.Context) {
	user := middleware.CurrentUser(c)
	list, err := h.approvalSvc.GetPendingList(c.Request.Context(), user.ID)
	if err != nil {
		utils.FailMsg(c, http.StatusInternalServerError, "查询待审批列表失败")
		return
	}
	utils.OK(c, list)
}

// Act 对指定审批节点执行审批操作
func (h *ApprovalHandler) Act(c *gin.Context) {
	user := middleware.CurrentUser(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.FailMsg(c, http.StatusBadRequest, "无效的审批 ID")
		return
	}
	var req models.ApprovalActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailMsg(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}
	if err := h.approvalSvc.Act(c.Request.Context(), id, user.ID, &req); err != nil {
		utils.FailMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.OK(c, gin.H{"message": "审批操作已提交"})
}

// GetLeaveApprovals 查询指定请假单的审批记录
func (h *ApprovalHandler) GetLeaveApprovals(c *gin.Context) {
	leaveID, err := strconv.ParseInt(c.Param("leave_id"), 10, 64)
	if err != nil {
		utils.FailMsg(c, http.StatusBadRequest, "无效的请假 ID")
		return
	}
	approvals, err := h.approvalSvc.GetLeaveApprovals(c.Request.Context(), leaveID)
	if err != nil {
		utils.FailMsg(c, http.StatusInternalServerError, "查询审批记录失败")
		return
	}
	utils.OK(c, approvals)
}
