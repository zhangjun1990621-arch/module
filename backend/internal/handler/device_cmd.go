package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"iot-platform/internal/model"
)

// DeviceCommandHandler 设备命令处理器（模拟 MQTT 指令下发）
type DeviceCommandHandler struct{}

func NewDeviceCommandHandler() *DeviceCommandHandler {
	return &DeviceCommandHandler{}
}

// PollDevice 召测设备（模拟）
func (h *DeviceCommandHandler) PollDevice(c *gin.Context) {
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
		fail(c, http.StatusNotFound, "设备不存在")
		return
	}

	var req struct {
		Items []string `json:"items"`
	}
	c.ShouldBindJSON(&req)

	// 模拟召测：更新设备状态为在线，刷新最后在线时间
	now := time.Now()
	pdb.Model(&device).Updates(map[string]interface{}{
		"status":    model.DeviceStatusOnline,
		"last_seen": now,
	})

	// 重新查询返回最新数据
	pdb.Where("id = ?", id).First(&device)

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true

	// 生成模拟实时数据（光伏逆变器典型参数）
	second := now.Second()
	realtimeData := gin.H{
		"ac": gin.H{
			"ph": 3,
			"v":  [3]float64{238.5 + float64(second%5), 239.2 + float64(second%4), 237.8 + float64(second%6)},
			"c":  [3]float64{15.6 + float64(second%3)*0.2, 15.8 + float64(second%3)*0.1, 15.4 + float64(second%3)*0.3},
			"p":  11000 + second*10,
			"q":  220 + second,
			"pf": 0.98,
			"f":  50.0 + float64(second%3)*0.01,
		},
		"dc": gin.H{
			"v": [3]float64{650.5 + float64(second%4), 648.2 + float64(second%3), 655.8 + float64(second%5)},
			"c": [3]float64{5.8 + float64(second%3)*0.1, 6.0 + float64(second%3)*0.15, 5.6 + float64(second%3)*0.2},
			"p": [3]float64{3772.9, 3889.2, 3672.5},
		},
		"cs": 28,
	}

	success(c, gin.H{
		"message":   "召测成功",
		"device":    device,
		"realtime":  realtimeData,
	})
}

// RebootDevice 重启设备（模拟）
func (h *DeviceCommandHandler) RebootDevice(c *gin.Context) {
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
		fail(c, http.StatusNotFound, "设备不存在")
		return
	}

	// 模拟重启：设备短暂离线后恢复
	pdb.Model(&device).Update("status", model.DeviceStatusOffline)

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, gin.H{"message": "重启指令已下发"})
}

// FactoryReset 恢复出厂设置（模拟）
func (h *DeviceCommandHandler) FactoryReset(c *gin.Context) {
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
		fail(c, http.StatusNotFound, "设备不存在")
		return
	}

	// 模拟恢复出厂
	pdb.Model(&device).Updates(map[string]interface{}{
		"status":   model.DeviceStatusOffline,
		"metadata": datatypes.JSON([]byte(fmt.Sprintf(`{"factory_reset": true, "reset_time": "%s"}`, time.Now().Format(time.RFC3339)))),
	})

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, gin.H{"message": "恢复出厂指令已下发"})
}

// ReportAck OTA准备（模拟下发上报确认帧）
func (h *DeviceCommandHandler) ReportAck(c *gin.Context) {
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
		fail(c, http.StatusNotFound, "设备不存在")
		return
	}

	// 模拟 OTA 准备：标记设备为 OTA 就绪
	pdb.Model(&device).Updates(map[string]interface{}{
		"status": model.DeviceStatusOnline,
	})

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, gin.H{"message": "OTA准备指令已下发，设备已进入升级准备状态"})
}

// SetDevice 设置参数（模拟）
func (h *DeviceCommandHandler) SetDevice(c *gin.Context) {
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
		fail(c, http.StatusNotFound, "设备不存在")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 模拟参数设置
	updates := map[string]interface{}{}
	if v, ok := req["activePower"]; ok {
		updates["metadata"] = datatypes.JSON([]byte(fmt.Sprintf(`{"active_power": %v}`, v)))
	}
	if len(updates) > 0 {
		pdb.Model(&device).Updates(updates)
	}

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true
	success(c, gin.H{"message": "参数设置已下发"})
}
