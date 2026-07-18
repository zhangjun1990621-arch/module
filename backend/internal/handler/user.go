package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"iot-platform/internal/middleware"
	"iot-platform/internal/model"
)

// UserHandler 用户管理处理器（仅超级管理员可调用）
type UserHandler struct {
	DB *gorm.DB
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// userListItem 用户列表项（不含密码）
type userListItem struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	Platforms string     `json:"platforms"`
	Status    string     `json:"status"`
	LastLogin *time.Time `json:"lastLogin"`
	CreatedAt time.Time  `json:"createdAt"`
}

// createUserRequest 创建用户请求体
type createUserRequest struct {
	Username  string `json:"username" binding:"required,min=2,max=64"`
	Password  string `json:"password" binding:"required,min=6,max=128"`
	Role      string `json:"role" binding:"required"`
	Platforms string `json:"platforms"`
	Status    string `json:"status"`
}

// updateUserRequest 更新用户请求体
type updateUserRequest struct {
	Role      *string `json:"role"`
	Platforms *string `json:"platforms"`
	Status    *string `json:"status"`
}

// resetPasswordRequest 重置密码请求体
type resetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// List 用户列表（支持分页、搜索、角色筛选）
func (h *UserHandler) List(c *gin.Context) {
	page, pageSize, offset := pagination(c)

	query := h.DB.Model(&model.User{})

	// 角色筛选
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	// 状态筛选
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	// 搜索（用户名模糊匹配）
	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		query = query.Where("username ILIKE ?", like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询用户总数失败: "+err.Error())
		return
	}

	var users []model.User
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询用户列表失败: "+err.Error())
		return
	}

	// 转换为列表项（排除密码字段）
	list := make([]userListItem, 0, len(users))
	for _, u := range users {
		list = append(list, userListItem{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			Platforms: u.Platforms,
			Status:    u.Status,
			LastLogin: u.LastLogin,
			CreatedAt: u.CreatedAt,
		})
	}

	pagedSuccess(c, list, total, page, pageSize)
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 校验角色
	if !isValidRole(req.Role) {
		fail(c, http.StatusBadRequest, "无效的角色: "+req.Role)
		return
	}

	// 超级管理员创建的非超管用户必须指定可访问平台
	if req.Role != model.RoleSuperAdmin && strings.TrimSpace(req.Platforms) == "" {
		fail(c, http.StatusBadRequest, "非超级管理员用户必须指定可访问平台")
		return
	}

	// 检查用户名是否已存在
	var count int64
	h.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		fail(c, http.StatusConflict, "用户名已存在")
		return
	}

	// 哈希密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, http.StatusInternalServerError, "密码加密失败")
		return
	}

	status := req.Status
	if status == "" {
		status = model.UserStatusActive
	}

	user := model.User{
		ID:        uuid.NewString(),
		Username:  req.Username,
		Password:  string(hashed),
		Role:      req.Role,
		Platforms: strings.TrimSpace(req.Platforms),
		Status:    status,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		fail(c, http.StatusInternalServerError, "创建用户失败: "+err.Error())
		return
	}

	success(c, userListItem{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Platforms: user.Platforms,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	})
}

// Update 更新用户信息（角色、可访问平台、状态）
func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetString(middleware.ContextKeyUserID)

	// 不允许修改自己
	if id == currentUserID {
		fail(c, http.StatusForbidden, "不能修改自己的账号信息")
		return
	}

	var user model.User
	if err := h.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "用户不存在")
			return
		}
		fail(c, http.StatusInternalServerError, "查询用户失败: "+err.Error())
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{}

	if req.Role != nil {
		if !isValidRole(*req.Role) {
			fail(c, http.StatusBadRequest, "无效的角色: "+*req.Role)
			return
		}
		// 如果要将目标用户降级为非超管，检查是否是最后一个超管
		if user.Role == model.RoleSuperAdmin && *req.Role != model.RoleSuperAdmin {
			var superAdminCount int64
			h.DB.Model(&model.User{}).Where("role = ? AND status = ?", model.RoleSuperAdmin, model.UserStatusActive).Count(&superAdminCount)
			if superAdminCount <= 1 {
				fail(c, http.StatusForbidden, "不能降级最后一个超级管理员")
				return
			}
		}
		updates["role"] = *req.Role
	}

	if req.Platforms != nil {
		// 非超管用户必须指定平台
		newRole := user.Role
		if req.Role != nil {
			newRole = *req.Role
		}
		if newRole != model.RoleSuperAdmin && strings.TrimSpace(*req.Platforms) == "" {
			fail(c, http.StatusBadRequest, "非超级管理员用户必须指定可访问平台")
			return
		}
		updates["platforms"] = strings.TrimSpace(*req.Platforms)
	}

	if req.Status != nil {
		// 如果要禁用超管，检查是否是最后一个活跃超管
		if user.Role == model.RoleSuperAdmin && *req.Status != model.UserStatusActive {
			var activeSuperAdminCount int64
			h.DB.Model(&model.User{}).Where("role = ? AND status = ?", model.RoleSuperAdmin, model.UserStatusActive).Count(&activeSuperAdminCount)
			if activeSuperAdminCount <= 1 {
				fail(c, http.StatusForbidden, "不能禁用最后一个活跃的超级管理员")
				return
			}
		}
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		fail(c, http.StatusBadRequest, "没有需要更新的字段")
		return
	}

	if err := h.DB.Model(&user).Updates(updates).Error; err != nil {
		fail(c, http.StatusInternalServerError, "更新用户失败: "+err.Error())
		return
	}

	// 重新查询返回最新数据
	h.DB.Where("id = ?", id).First(&user)
	success(c, userListItem{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Platforms: user.Platforms,
		Status:    user.Status,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
	})
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetString(middleware.ContextKeyUserID)

	// 不允许删除自己
	if id == currentUserID {
		fail(c, http.StatusForbidden, "不能删除自己的账号")
		return
	}

	var user model.User
	if err := h.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "用户不存在")
			return
		}
		fail(c, http.StatusInternalServerError, "查询用户失败: "+err.Error())
		return
	}

	// 如果要删除超管，检查是否是最后一个活跃超管
	if user.Role == model.RoleSuperAdmin {
		var activeSuperAdminCount int64
		h.DB.Model(&model.User{}).Where("role = ? AND status = ?", model.RoleSuperAdmin, model.UserStatusActive).Count(&activeSuperAdminCount)
		if activeSuperAdminCount <= 1 {
			fail(c, http.StatusForbidden, "不能删除最后一个超级管理员")
			return
		}
	}

	if err := h.DB.Delete(&user).Error; err != nil {
		fail(c, http.StatusInternalServerError, "删除用户失败: "+err.Error())
		return
	}

	success(c, gin.H{"message": "用户已删除"})
}

// ResetPassword 重置用户密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 检查用户是否存在
	var count int64
	h.DB.Model(&model.User{}).Where("id = ?", id).Count(&count)
	if count == 0 {
		fail(c, http.StatusNotFound, "用户不存在")
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(c, http.StatusInternalServerError, "密码加密失败")
		return
	}

	if err := h.DB.Model(&model.User{}).Where("id = ?", id).Update("password", string(hashed)).Error; err != nil {
		fail(c, http.StatusInternalServerError, "重置密码失败: "+err.Error())
		return
	}

	success(c, gin.H{"message": "密码已重置"})
}

// isValidRole 校验角色是否合法
func isValidRole(role string) bool {
	return role == model.RoleSuperAdmin || role == model.RoleAdmin || role == model.RoleViewer
}
