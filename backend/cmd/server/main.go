package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"iot-platform/internal/config"
	"iot-platform/internal/database"
	"iot-platform/internal/handler"
	"iot-platform/internal/middleware"
	"iot-platform/internal/model"
	"iot-platform/internal/service"
)

func main() {
	// 1. 加载配置
	configPath := getEnv("CONFIG_PATH", "config/config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 2. 连接数据库并自动迁移
	if err := database.Init(cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 3. 初始化默认数据
	if err := initDefaultData(database.DB); err != nil {
		log.Printf("初始化默认数据警告: %v", err)
	}

	// 4. 初始化服务与处理器
	platformSvc := service.NewPlatformService(database.DB)
	authHandler := handler.NewAuthHandler(database.DB, cfg.JWT.Secret, cfg.JWT.Expire)
	platformHandler := handler.NewPlatformHandler(platformSvc)
	deviceHandler := handler.NewDeviceHandler()
	alarmHandler := handler.NewAlarmHandler()
	dashboardHandler := handler.NewDashboardHandler()
	userHandler := handler.NewUserHandler(database.DB)
	otaHandler := handler.NewOTAHandler()

	// 5. 注册路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(corsMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 认证路由(无需认证，profile/logout 需认证)
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/profile", middleware.AuthMiddleware(cfg.JWT.Secret), authHandler.Profile)
		auth.POST("/logout", middleware.AuthMiddleware(cfg.JWT.Secret), authHandler.Logout)
	}

	// 平台管理路由(需认证)
	platformsAPI := r.Group("/api/platforms")
	platformsAPI.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		platformsAPI.GET("", platformHandler.List)
		platformsAPI.GET("/", platformHandler.List)
		platformsAPI.GET("/:id", platformHandler.Get)
		platformsAPI.POST("", middleware.RequireRole(model.RoleSuperAdmin), platformHandler.Create)
		platformsAPI.PUT("/:id", platformHandler.Update)
		platformsAPI.DELETE("/:id", middleware.RequireRole(model.RoleSuperAdmin), platformHandler.Delete)
	}

	// 用户管理路由(需认证 + 超级管理员权限)
	usersAPI := r.Group("/api/users")
	usersAPI.Use(middleware.AuthMiddleware(cfg.JWT.Secret), middleware.RequireRole(model.RoleSuperAdmin))
	{
		usersAPI.GET("", userHandler.List)
		usersAPI.GET("/", userHandler.List)
		usersAPI.POST("", userHandler.Create)
		usersAPI.PUT("/:id", userHandler.Update)
		usersAPI.DELETE("/:id", userHandler.Delete)
		usersAPI.PUT("/:id/password", userHandler.ResetPassword)
	}

	// 平台专属业务路由(需认证 + 平台访问权限校验)
	platformAPI := r.Group("/api/:platform")
	platformAPI.Use(middleware.AuthMiddleware(cfg.JWT.Secret), middleware.PlatformAccess())
	{
		// 设备 CRUD
		platformAPI.GET("/devices", deviceHandler.List)
		platformAPI.GET("/devices/:id", deviceHandler.Get)
		platformAPI.POST("/devices", deviceHandler.Create)
		platformAPI.PUT("/devices/:id", deviceHandler.Update)
		platformAPI.DELETE("/devices/:id", deviceHandler.Delete)

		// 告警查询与处理
		platformAPI.GET("/alarms", alarmHandler.List)
		platformAPI.PUT("/alarms/:id/resolve", alarmHandler.Resolve)

		// 仪表盘
		platformAPI.GET("/dashboard", dashboardHandler.Get)

		// OTA 升级管理
		otaAPI := platformAPI.Group("/ota")
		{
			// 固件管理
			otaAPI.GET("/firmwares", otaHandler.ListFirmwares)
			otaAPI.POST("/firmwares", otaHandler.UploadFirmware)
			otaAPI.DELETE("/firmwares/:id", otaHandler.DeleteFirmware)

			// 升级任务
			otaAPI.GET("/tasks", otaHandler.ListTasks)
			otaAPI.GET("/tasks/:id", otaHandler.GetTask)
			otaAPI.POST("/tasks", otaHandler.CreateTask)
			otaAPI.DELETE("/tasks/:id", otaHandler.DeleteTask)
			otaAPI.POST("/tasks/:id/pause", otaHandler.PauseTask)
			otaAPI.POST("/tasks/:id/resume", otaHandler.ResumeTask)
			otaAPI.POST("/tasks/:id/cancel", otaHandler.CancelTask)
			otaAPI.POST("/tasks/:id/complete", otaHandler.CompleteTask)
			otaAPI.POST("/tasks/:id/retry", otaHandler.RetryFailedDevices)
		}
	}

	// 6. 启动 HTTP 服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("IoT 平台后端服务启动，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// initDefaultData 初始化默认数据：若表为空则插入示例平台与管理员用户
func initDefaultData(db *gorm.DB) error {
	// 初始化管理员用户
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	if userCount == 0 {
		hashed, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("生成密码哈希失败: %w", err)
		}
		admin := model.User{
			ID:        "admin",
			Username:  "admin",
			Password:  string(hashed),
			Role:      model.RoleSuperAdmin,
			Platforms: "aluminum,pv,power,water",
			Status:    model.UserStatusActive,
		}
		if err := db.Create(&admin).Error; err != nil {
			return fmt.Errorf("创建管理员失败: %w", err)
		}
		log.Println("已创建默认管理员: admin / admin123")
	}

	// 初始化示例平台
	var platformCount int64
	db.Model(&model.Platform{}).Count(&platformCount)
	if platformCount == 0 {
		platformSvc := service.NewPlatformService(db)
		for _, p := range buildDefaultPlatforms() {
			_, err := platformSvc.CreatePlatform(p)
			if err != nil {
				log.Printf("创建平台 %s 失败: %v", p.ID, err)
				continue
			}
			log.Printf("已创建示例平台: %s (%s)", p.ID, p.Name)
		}
	}

	return nil
}

// buildDefaultPlatforms 构建铝厂与光伏两个示例平台的默认配置
func buildDefaultPlatforms() []service.CreatePlatformInput {
	return []service.CreatePlatformInput{
		{
			ID:        "aluminum",
			Name:      "铝厂云控平台",
			Icon:      "🏭",
			Schema:    "schema_aluminum",
			Status:    "active",
			SortOrder: 1,
			Config: &model.PlatformConfig{
				NavItems: []model.NavItem{
					{Path: "/aluminum/dashboard", Label: "仪表盘", Icon: "Odometer"},
					{Path: "/aluminum/devices", Label: "设备列表", Icon: "Monitor"},
					{Path: "/aluminum/alarms", Label: "告警管理", Icon: "Bell"},
				},
				Pages: map[string]model.PageDef{
					"dashboard": {Type: "dashboard", Title: "铝厂云控仪表盘", API: "dashboard"},
					"devices": {
						Type:  "table",
						Title: "设备列表",
						API:   "devices",
						Columns: []model.ColumnDef{
							{Field: "deviceId", Label: "设备ID", Type: "text", Width: 160},
							{Field: "name", Label: "设备名称", Type: "text", Width: 180},
							{Field: "stationId", Label: "站点", Type: "text", Width: 120},
							{Field: "status", Label: "状态", Type: "tag", Width: 100, Options: map[string]string{"online": "在线", "offline": "离线", "alarm": "告警"}},
							{Field: "lastSeen", Label: "最后在线", Type: "text", Width: 180},
						},
					},
					"alarms": {
						Type:  "table",
						Title: "告警管理",
						API:   "alarms",
						Columns: []model.ColumnDef{
							{Field: "deviceName", Label: "设备", Type: "text", Width: 160},
							{Field: "level", Label: "等级", Type: "tag", Width: 100, Options: map[string]string{"critical": "严重", "warning": "警告", "info": "信息"}},
							{Field: "type", Label: "类型", Type: "text", Width: 140},
							{Field: "detail", Label: "详情", Type: "text"},
							{Field: "status", Label: "状态", Type: "tag", Width: 100, Options: map[string]string{"active": "未处理", "resolved": "已处理"}},
							{Field: "occurredAt", Label: "发生时间", Type: "text", Width: 180},
						},
					},
				},
			},
		},
		{
			ID:        "pv",
			Name:      "光伏运维平台",
			Icon:      "☀️",
			Schema:    "schema_pv",
			Status:    "active",
			SortOrder: 2,
			Config: &model.PlatformConfig{
				NavItems: []model.NavItem{
					{Path: "/pv/dashboard", Label: "仪表盘", Icon: "Odometer"},
					{Path: "/pv/devices", Label: "设备管理", Icon: "Cpu"},
					{Path: "/pv/alarms", Label: "事件告警", Icon: "Warning"},
				},
				Pages: map[string]model.PageDef{
					"dashboard": {Type: "dashboard", Title: "光伏运维仪表盘", API: "dashboard"},
					"devices": {
						Type:  "table",
						Title: "设备管理",
						API:   "devices",
						Columns: []model.ColumnDef{
							{Field: "deviceId", Label: "设备ID", Type: "text", Width: 160},
							{Field: "name", Label: "设备名称", Type: "text", Width: 180},
							{Field: "stationId", Label: "电站", Type: "text", Width: 120},
							{Field: "status", Label: "状态", Type: "tag", Width: 100, Options: map[string]string{"online": "在线", "offline": "离线", "alarm": "告警"}},
							{Field: "lastSeen", Label: "最后在线", Type: "text", Width: 180},
						},
					},
					"alarms": {
						Type:  "table",
						Title: "事件告警",
						API:   "alarms",
						Columns: []model.ColumnDef{
							{Field: "deviceName", Label: "设备", Type: "text", Width: 160},
							{Field: "level", Label: "等级", Type: "tag", Width: 100, Options: map[string]string{"critical": "严重", "warning": "警告", "info": "信息"}},
							{Field: "type", Label: "类型", Type: "text", Width: 140},
							{Field: "detail", Label: "详情", Type: "text"},
							{Field: "status", Label: "状态", Type: "tag", Width: 100, Options: map[string]string{"active": "未处理", "resolved": "已处理"}},
							{Field: "occurredAt", Label: "发生时间", Type: "text", Width: 180},
						},
					},
				},
			},
		},
	}
}

// corsMiddleware 简单 CORS 中间件，允许前端跨域访问
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// getEnv 读取环境变量，不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return defaultValue
}
