package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 API 响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedResponse 带分页信息的响应
type PagedResponse struct {
	Response
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// OK 返回 200 成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// OKPaged 返回带分页的成功响应
func OKPaged(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, PagedResponse{
		Response: Response{Code: 0, Message: "success", Data: data},
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Fail 返回业务失败响应
func Fail(c *gin.Context, appErr *AppError) {
	c.JSON(appErr.Code, Response{
		Code:    appErr.Code,
		Message: appErr.Message,
	})
}

// FailMsg 返回指定状态码和消息的失败响应
func FailMsg(c *gin.Context, status int, msg string) {
	c.JSON(status, Response{
		Code:    status,
		Message: msg,
	})
}
