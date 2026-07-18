package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedData 分页数据结构
type PagedData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// success 返回成功响应
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// pagedSuccess 返回分页成功响应
func pagedSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data: PagedData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// fail 返回失败响应
func fail(c *gin.Context, status int, msg string) {
	c.JSON(status, APIResponse{
		Code:    status,
		Message: msg,
	})
}

// pagination 从查询参数解析分页参数
func pagination(c *gin.Context) (page, pageSize, offset int) {
	page = 1
	pageSize = 20

	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("pageSize", "20")); err == nil && ps > 0 && ps <= 200 {
		pageSize = ps
	}
	offset = (page - 1) * pageSize
	return
}
