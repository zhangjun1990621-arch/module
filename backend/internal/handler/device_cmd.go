package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"iot-platform/internal/middleware"
	"iot-platform/internal/model"
	mqttClient "iot-platform/internal/mqtt"
	"iot-platform/internal/service"
)

// DeviceCommandHandler 设备命令处理器
type DeviceCommandHandler struct {
	publisher *mqttClient.Publisher
	tokenMgr  *service.TokenManager
}

// NewDeviceCommandHandler 创建设备命令处理器
func NewDeviceCommandHandler(publisher *mqttClient.Publisher, tokenMgr *service.TokenManager) *DeviceCommandHandler {
	return &DeviceCommandHandler{
		publisher: publisher,
		tokenMgr:  tokenMgr,
	}
}

// PollDevice 召测设备
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
	if err := pdb.Where("id = ? OR device_id = ?", id, id).First(&device).Error; err != nil {
		fail(c, http.StatusNotFound, "设备不存在")
		return
	}

	// 更新设备状态为在线
	now := time.Now()
	pdb.Model(&device).Updates(map[string]interface{}{
		"status":    model.DeviceStatusOnline,
		"last_seen": now,
	})
	pdb.Where("id = ?", device.ID).First(&device)

	if err := pdb.Commit().Error; err != nil {
		fail(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}
	committed = true

	// 如果有 MQTT publisher，通过 MQTT 下发召测命令
	var realtimeData interface{}
	var mqttResponded bool
	if h.publisher != nil && h.tokenMgr != nil {
		token := int(h.tokenMgr.NextToken())
		items := []string{"ac", "dc", "sw", "hw", "md", "sn", "cp", "fw"}

		// 注册等待响应
		respCh := h.tokenMgr.Register(device.DeviceID, uint16(token))

		// 下发召测命令
		if err := h.publisher.GetCommand(device.DeviceID, token, items); err != nil {
			success(c, gin.H{
				"message":  "召测指令下发失败，设备状态已更新",
				"device":   device,
				"realtime": nil,
			})
			return
		}

		// 等待响应（最多 5 秒）
		select {
		case resp := <-respCh:
			if resp.Data != nil {
				realtimeData = gin.H{
					"ac":         resp.Data.AC,
					"dc":         resp.Data.DC,
					"software":   resp.Data.Software,
					"hardware":   resp.Data.Hardware,
					"model":      resp.Data.Model,
					"serialNo":   resp.Data.SerialNo,
					"capacity":   resp.Data.Capacity,
					"firmware":   resp.Data.Firmware,
					"deviceType": resp.Data.DeviceType,
				}
				mqttResponded = true
			}
		case <-time.After(5 * time.Second):
			// 超时，设备未响应
		}
	}

	if !mqttResponded {
		success(c, gin.H{
			"message":  "召测指令已下发，设备状态已更新（设备未返回实时数据）",
			"device":   device,
			"realtime": nil,
		})
		return
	}

	success(c, gin.H{
		"message":  "召测成功，设备已返回实时数据",
		"device":   device,
		"realtime": realtimeData,
	})
}

// RebootDevice 重启设备
func (h *DeviceCommandHandler) RebootDevice(c *gin.Context) {
	id := c.Param("id")

	if h.publisher != nil {
		token := int(h.tokenMgr.NextToken())
		if err := h.publisher.Reboot(id, token); err != nil {
			fail(c, http.StatusInternalServerError, "重启指令下发失败: "+err.Error())
			return
		}
	}

	success(c, gin.H{"message": "重启指令已下发"})
}

// FactoryReset 恢复出厂设置
func (h *DeviceCommandHandler) FactoryReset(c *gin.Context) {
	id := c.Param("id")

	if h.publisher != nil {
		token := int(h.tokenMgr.NextToken())
		if err := h.publisher.FactoryReset(id, token); err != nil {
			fail(c, http.StatusInternalServerError, "恢复出厂指令下发失败: "+err.Error())
			return
		}
	}

	success(c, gin.H{"message": "恢复出厂指令已下发"})
}

// ReportAck OTA 准备
func (h *DeviceCommandHandler) ReportAck(c *gin.Context) {
	id := c.Param("id")

	if h.publisher != nil {
		if err := h.publisher.ReportAck(id); err != nil {
			fail(c, http.StatusInternalServerError, "OTA准备指令下发失败: "+err.Error())
			return
		}
	}

	success(c, gin.H{"message": "OTA准备指令已下发，设备已进入升级准备状态"})
}

// SetDevice 设置参数
func (h *DeviceCommandHandler) SetDevice(c *gin.Context) {
	id := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if h.publisher != nil {
		token := int(h.tokenMgr.NextToken())
		if err := h.publisher.SetCommand(id, token, req); err != nil {
			fail(c, http.StatusInternalServerError, "参数设置指令下发失败: "+err.Error())
			return
		}
	}

	success(c, gin.H{"message": "参数设置已下发"})
}

// 保留 unused import 引用
var _ = datatypes.JSON{}
var _ = fmt.Sprintf
var _ = middleware.AuthMiddleware
