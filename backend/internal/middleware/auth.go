package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"iot-platform/internal/model"
)

// Context 键名常量
const (
	ContextKeyUserID    = "userID"
	ContextKeyUsername  = "username"
	ContextKeyRole      = "role"
	ContextKeyPlatforms = "platforms"
)

// Claims JWT 自定义声明
type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Platforms string `json:"platforms"`
	jwt.RegisteredClaims
}

// GenerateToken 根据用户信息生成 JWT
func GenerateToken(secret string, expireHours int, user *model.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Platforms: user.Platforms,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireHours) * time.Hour)),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证 JWT
func ParseToken(secret string, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// AuthMiddleware JWT 认证中间件：解析 Authorization 头并将用户信息注入 gin.Context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少认证信息"})
			return
		}

		// 期望格式 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			return
		}

		claims, err := ParseToken(jwtSecret, strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的令牌"})
			return
		}

		// 注入用户信息到上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyPlatforms, claims.Platforms)

		c.Next()
	}
}

// RequireRole 角色权限校验中间件，仅允许指定角色通过
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(ContextKeyRole)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足"})
	}
}

// PlatformAccess 平台访问权限校验中间件。
// super_admin 可访问所有平台；其他角色需 platforms 列表中包含 :platform 参数。
func PlatformAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(ContextKeyRole)
		if role == model.RoleSuperAdmin {
			c.Next()
			return
		}

		platform := c.Param("platform")
		platforms := c.GetString(ContextKeyPlatforms)
		if !platformAllowed(platforms, platform) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权访问该平台"})
			return
		}
		c.Next()
	}
}

// platformAllowed 判断逗号分隔的平台列表是否包含目标平台
func platformAllowed(platformsCSV string, target string) bool {
	if platformsCSV == "" {
		return false
	}
	for _, p := range strings.Split(platformsCSV, ",") {
		if strings.TrimSpace(p) == target {
			return true
		}
	}
	return false
}
