package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"iot-platform/internal/database"
	"iot-platform/internal/model"
)

// DeviceHandler 通用设备处理器，根据 :platform 参数路由到对应 schema 的 devices 表
type DeviceHandler struct{}

// NewDeviceHandler 创建设备处理器
func NewDeviceHandler() *DeviceHandler {
	return &DeviceHandler{}
}

// getPlatformTx 获取平台专属事务 DB(已设置 search_path)，失败时已写入错误响应
func getPlatformTx(c *gin.Context) (*gorm.DB, bool) {
	platform := c.Param("platform")
	pdb, err := database.GetPlatformDB(platform)
	if err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return nil, false
	}
	return pdb, true
}

// List 设备列表(支持分页、搜索、状态筛选)
func (h *DeviceHandler) List(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	page, pageSize, offset := pagination(c)

	query := pdb.Model(&model.Device{})

	// 平台内设备可按 platform_id 过滤(平台 schema 内一般已限定，保留以兼容 public 表)
	if pid := c.Query("platformId"); pid != "" {
		query = query.Where("platform_id = ?", pid)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR device_id ILIKE ? OR station_id ILIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询设备总数失败: "+err.Error())
		return
	}

	var devices []model.Device
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&devices).Error; err != nil {
		fail(c, http.StatusInternalServerError, "查询设备列表失败: "+err.Error())
		return
	}
	if devices == nil {
		devices = make([]model.Device, 0)
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败")
		return
	}
	committed = true
	pagedSuccess(c, devices, total, page, pageSize)
}

// Get 设备详情
func (h *DeviceHandler) Get(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var device model.Device
	if err := pdb.Where("id = ?", id).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "设备不存在")
			return
		}
		fail(c, http.StatusInternalServerError, "查询设备失败: "+err.Error())
		return
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败")
		return
	}
	committed = true
	success(c, device)
}

// createDeviceRequest 创建设备请求体
type createDeviceRequest struct {
	DeviceID  string            `json:"deviceId" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	StationID string            `json:"stationId"`
	Status    string            `json:"status"`
	Metadata  datatypes.JSON   `json:"metadata"`
}

// Create 创建设备
func (h *DeviceHandler) Create(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	var req createDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	platform := c.Param("platform")
	status := req.Status
	if status == "" {
		status = model.DeviceStatusOffline
	}
	if len(req.Metadata) == 0 {
		req.Metadata = datatypes.JSON([]byte("{}"))
	}

	device := model.Device{
		ID:         uuid.NewString(),
		PlatformID: platform,
		DeviceID:   req.DeviceID,
		Name:       req.Name,
		StationID:  req.StationID,
		Status:     status,
		Metadata:   req.Metadata,
	}

	if err := pdb.Create(&device).Error; err != nil {
		fail(c, http.StatusInternalServerError, "创建设备失败: "+err.Error())
		return
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败")
		return
	}
	committed = true
	c.JSON(http.StatusCreated, APIResponse{Code: 0, Message: "success", Data: device})
}

// updateDeviceRequest 更新设备请求体
type updateDeviceRequest struct {
	DeviceID  *string          `json:"deviceId"`
	Name      *string          `json:"name"`
	StationID *string          `json:"stationId"`
	Status    *string          `json:"status"`
	Metadata  *datatypes.JSON  `json:"metadata"`
	LastSeen  *time.Time       `json:"lastSeen"`
}

// Update 更新设备
func (h *DeviceHandler) Update(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	var device model.Device
	if err := pdb.Where("id = ?", id).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fail(c, http.StatusNotFound, "设备不存在")
			return
		}
		fail(c, http.StatusInternalServerError, "查询设备失败: "+err.Error())
		return
	}

	var req updateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.DeviceID != nil {
		updates["device_id"] = *req.DeviceID
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.StationID != nil {
		updates["station_id"] = *req.StationID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Metadata != nil {
		updates["metadata"] = *req.Metadata
	}
	if req.LastSeen != nil {
		updates["last_seen"] = *req.LastSeen
	}

	if len(updates) > 0 {
		if err := pdb.Model(&device).Updates(updates).Error; err != nil {
			fail(c, http.StatusInternalServerError, "更新设备失败: "+err.Error())
			return
		}
	}

	// 重新查询返回最新数据
	pdb.Where("id = ?", id).First(&device)

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败")
		return
	}
	committed = true
	success(c, device)
}

// Delete 删除设备
func (h *DeviceHandler) Delete(c *gin.Context) {
	pdb, ok := getPlatformTx(c)
	if !ok {
		return
	}
	committed := false
	defer func() {
		if !committed {
			pdb.Rollback()
		}
	}()

	id := c.Param("id")
	result := pdb.Where("id = ?", id).Delete(&model.Device{})
	if result.Error != nil {
		fail(c, http.StatusInternalServerError, "删除设备失败: "+result.Error.Error())
		return
	}
	if result.RowsAffected == 0 {
		fail(c, http.StatusNotFound, "设备不存在")
		return
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败")
		return
	}
	committed = true
	success(c, gin.H{"message": "设备已删除"})
}
