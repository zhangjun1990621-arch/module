package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"iot-platform/internal/middleware"
	"iot-platform/internal/model"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	DB         *gorm.DB
	JWTSecret  string
	JWTExpire  int
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *gorm.DB, jwtSecret string, jwtExpire int) *AuthHandler {
	return &AuthHandler{
		DB:        db,
		JWTSecret: jwtSecret,
		JWTExpire: jwtExpire,
	}
}

// loginRequest 登录请求体
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// loginResponse 登录响应体
type loginResponse struct {
	Token string      `json:"token"`
	User  userInfo    `json:"user"`
}

type userInfo struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	Platforms string     `json:"platforms"`
	Status    string     `json:"status"`
	LastLogin *time.Time `json:"lastLogin"`
}

// Login 用户登录，校验用户名密码并返回 JWT
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 查询用户
	var user model.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		fail(c, http.StatusInternalServerError, "查询用户失败")
		return
	}

	// 校验状态
	if user.Status != model.UserStatusActive {
		fail(c, http.StatusForbidden, "账号已被禁用")
		return
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		fail(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 更新最后登录时间
	now := time.Now()
	h.DB.Model(&user).Update("last_login", now)

	// 生成 JWT
	token, err := middleware.GenerateToken(h.JWTSecret, h.JWTExpire, &user)
	if err != nil {
		fail(c, http.StatusInternalServerError, "生成令牌失败")
		return
	}

	success(c, loginResponse{
		Token: token,
		User: userInfo{
			ID:        user.ID,
			Username:  user.Username,
			Role:      user.Role,
			Platforms: user.Platforms,
			Status:    user.Status,
			LastLogin: &now,
		},
	})
}

// Profile 获取当前登录用户信息
func (h *AuthHandler) Profile(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	if userID == "" {
		fail(c, http.StatusUnauthorized, "未认证")
		return
	}

	var user model.User
	if err := h.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		fail(c, http.StatusNotFound, "用户不存在")
		return
	}

	success(c, userInfo{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Platforms: user.Platforms,
		Status:    user.Status,
		LastLogin: user.LastLogin,
	})
}

// Logout 登出。JWT 为无状态令牌，客户端清除即可，服务端仅返回成功。
func (h *AuthHandler) Logout(c *gin.Context) {
	success(c, gin.H{"message": "已登出"})
}
