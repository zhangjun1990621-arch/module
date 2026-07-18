package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"iot-platform/internal/middleware"
	"iot-platform/internal/model"
	"iot-platform/internal/service"
)

// PlatformHandler 平台处理器
type PlatformHandler struct {
	svc *service.PlatformService
}

// NewPlatformHandler 创建平台处理器
func NewPlatformHandler(svc *service.PlatformService) *PlatformHandler {
	return &PlatformHandler{svc: svc}
}

// List 返回所有 active 平台列表(含 config JSON)
func (h *PlatformHandler) List(c *gin.Context) {
	platforms, err := h.svc.GetActivePlatforms()
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询平台列表失败: "+err.Error())
		return
	}
	success(c, platforms)
}

// Get 返回单个平台详情
func (h *PlatformHandler) Get(c *gin.Context) {
	id := c.Param("id")
	platform, err := h.svc.GetPlatformByID(id)
	if err != nil {
		fail(c, http.StatusNotFound, "平台不存在")
		return
	}
	success(c, platform)
}

// createPlatformRequest 创建平台请求体
type createPlatformRequest struct {
	ID        string                 `json:"id" binding:"required"`
	Name      string                 `json:"name" binding:"required"`
	Icon      string                 `json:"icon"`
	Schema    string                 `json:"schema" binding:"required"`
	Config    *model.PlatformConfig  `json:"config"`
	Status    string                 `json:"status"`
	SortOrder int                    `json:"sortOrder"`
}

// Create 创建新平台(super_admin only)，自动创建 schema 与表
func (h *PlatformHandler) Create(c *gin.Context) {
	var req createPlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 规范化 schema 名：默认根据 ID 生成 schema_xxx
	schemaName := strings.TrimSpace(req.Schema)
	if schemaName == "" {
		schemaName = "schema_" + req.ID
	}

	in := service.CreatePlatformInput{
		ID:        req.ID,
		Name:      req.Name,
		Icon:      req.Icon,
		Schema:    schemaName,
		Config:    req.Config,
		Status:    req.Status,
		SortOrder: req.SortOrder,
	}

	platform, err := h.svc.CreatePlatform(in)
	if err != nil {
		fail(c, http.StatusBadRequest, "创建平台失败: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Code:    0,
		Message: "平台创建成功",
		Data:    platform,
	})
}

// updatePlatformRequest 更新平台请求体
type updatePlatformRequest struct {
	Name      *string                `json:"name"`
	Icon      *string                `json:"icon"`
	Config    *model.PlatformConfig  `json:"config"`
	Status    *string                `json:"status"`
	SortOrder *int                   `json:"sortOrder"`
}

// Update 更新平台信息
func (h *PlatformHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req updatePlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 仅 super_admin 可修改状态
	if req.Status != nil {
		role := c.GetString(middleware.ContextKeyRole)
		if role != model.RoleSuperAdmin {
			fail(c, http.StatusForbidden, "仅超级管理员可修改平台状态")
			return
		}
	}

	in := service.UpdatePlatformInput{
		Name:      req.Name,
		Icon:      req.Icon,
		Config:    req.Config,
		Status:    req.Status,
		SortOrder: req.SortOrder,
	}

	platform, err := h.svc.UpdatePlatform(id, in)
	if err != nil {
		fail(c, http.StatusInternalServerError, "更新平台失败: "+err.Error())
		return
	}
	success(c, platform)
}

// Delete 删除平台(super_admin only)
func (h *PlatformHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.DeletePlatform(id); err != nil {
		fail(c, http.StatusInternalServerError, "删除平台失败: "+err.Error())
		return
	}
	success(c, gin.H{"message": "平台已删除"})
}
