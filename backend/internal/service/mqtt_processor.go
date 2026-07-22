package service

import (
	"encoding/json"
	"log"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"iot-platform/internal/database"
	"iot-platform/internal/mqtt"
	"iot-platform/internal/model"
)

// MQTTEventProcessor MQTT 事件处理器
type MQTTEventProcessor struct {
	db        *gorm.DB
	publisher *mqtt.Publisher
	tokenMgr  *TokenManager
}

// NewMQTTEventProcessor 创建事件处理器
func NewMQTTEventProcessor(db *gorm.DB, tokenMgr *TokenManager) *MQTTEventProcessor {
	return &MQTTEventProcessor{
		db:       db,
		tokenMgr: tokenMgr,
	}
}

// SetPublisher 注入 MQTT 发布器
func (p *MQTTEventProcessor) SetPublisher(pub *mqtt.Publisher) {
	p.publisher = pub
}

// OnReport 处理设备上报消息
func (p *MQTTEventProcessor) OnReport(deviceID string, msg *mqtt.ReportMessage) {
	log.Printf("Processing report from %s: token=%d", deviceID, msg.Token)

	// 使用 pv 平台的数据库
	pdb := database.GetPlatformDBRaw("pv")
	if pdb == nil {
		log.Printf("Failed to get pv platform DB")
		return
	}

	device, isNewOnline, err := p.ensureDeviceExists(pdb, deviceID, msg.On, msg.Timestamp)
	if err != nil {
		log.Printf("Failed to ensure device exists: %v", err)
		return
	}

	if isNewOnline && msg.On != nil {
		log.Printf("Device %s online: sw=%s, hw=%s, cs=%d",
			device.ID, msg.On.Software, msg.On.Hardware, msg.On.SignalStrength)
	}

	if msg.HB != nil {
		p.handleHeartbeat(pdb, device, msg)
	}

	// 回复上报确认帧
	p.sendReportConfirm(deviceID, msg)
}

// sendReportConfirm 回复确认帧
func (p *MQTTEventProcessor) sendReportConfirm(deviceID string, msg *mqtt.ReportMessage) {
	if p.publisher == nil {
		return
	}

	fields := make(map[string]int)
	if msg.On != nil {
		fields["on"] = 0
	}
	if msg.Off != nil {
		fields["off"] = 0
	}
	if msg.HB != nil {
		fields["hb"] = 0
	}
	if msg.EOV != nil {
		fields["eov"] = 0
	}
	if msg.EOVR != nil {
		fields["eov_r"] = 0
	}
	if msg.EUV != nil {
		fields["euv"] = 0
	}
	if msg.EUVR != nil {
		fields["euv_r"] = 0
	}
	if msg.ELC != nil {
		fields["elc"] = 0
	}

	if len(fields) == 0 {
		fields["on"] = 0
	}

	if err := p.publisher.ReportConfirm(deviceID, msg.Token, fields); err != nil {
		log.Printf("Failed to send report confirm to %s: %v", deviceID, err)
	}
}

// ensureDeviceExists 确保设备存在并更新在线状态
func (p *MQTTEventProcessor) ensureDeviceExists(pdb *gorm.DB, deviceID string, onEvent *mqtt.OnlineEvent, timestamp int64) (*model.Device, bool, error) {
	var device model.Device
	err := pdb.Where("id = ? OR device_id = ?", deviceID, deviceID).First(&device).Error

	deviceTime := time.Now()
	if timestamp > 0 {
		deviceTime = time.Unix(timestamp, 0)
	}

	if err == gorm.ErrRecordNotFound {
		// 设备不存在，自动注册
		device = model.Device{
			ID:         deviceID,
			PlatformID: "pv",
			DeviceID:   deviceID,
			Name:       "设备-" + deviceID,
			Status:     model.DeviceStatusOnline,
			LastSeen:   &deviceTime,
		}
		if onEvent != nil {
			device.Metadata = mapToJSON(onEvent)
		}
		if err := pdb.Create(&device).Error; err != nil {
			return nil, false, err
		}
		log.Printf("Auto-registered new device: %s", deviceID)
		return &device, true, nil
	}

	if err != nil {
		return nil, false, err
	}

	wasOffline := device.Status != model.DeviceStatusOnline

	updates := map[string]interface{}{
		"status":    model.DeviceStatusOnline,
		"last_seen": deviceTime,
	}
	if onEvent != nil {
		updates["metadata"] = mapToJSON(onEvent)
	}
	pdb.Model(&device).Updates(updates)

	return &device, wasOffline, nil
}

// handleHeartbeat 处理心跳
func (p *MQTTEventProcessor) handleHeartbeat(pdb *gorm.DB, device *model.Device, msg *mqtt.ReportMessage) {
	now := time.Now()
	pdb.Model(&device).Updates(map[string]interface{}{
		"status":    model.DeviceStatusOnline,
		"last_seen": &now,
	})
}

// OnGetResp 处理召测响应
func (p *MQTTEventProcessor) OnGetResp(deviceID string, msg *mqtt.GetResponse) {
	log.Printf("Get response from %s: token=%d", deviceID, msg.Token)
	p.tokenMgr.Resolve(deviceID, uint16(msg.Token), MQTTResponse{
		Token: msg.Token,
		Data:  msg,
	})
}

// OnSetResp 处理设置响应
func (p *MQTTEventProcessor) OnSetResp(deviceID string, token int, confirm int) {
	log.Printf("Set response from %s: token=%d, confirm=%d", deviceID, token, confirm)
	p.tokenMgr.Resolve(deviceID, uint16(token), MQTTResponse{
		Token:   token,
		Confirm: confirm,
	})
}

// OnActionResp 处理动作响应
func (p *MQTTEventProcessor) OnActionResp(deviceID string, token int, confirm int) {
	log.Printf("Action response from %s: token=%d, confirm=%d", deviceID, token, confirm)
	p.tokenMgr.Resolve(deviceID, uint16(token), MQTTResponse{
		Token:   token,
		Confirm: confirm,
	})
}

// mapToJSON 将 OnlineEvent 转为 datatypes.JSON
func mapToJSON(on *mqtt.OnlineEvent) datatypes.JSON {
	data, _ := json.Marshal(map[string]interface{}{
		"software":       on.Software,
		"hardware":       on.Hardware,
		"signalStrength": on.SignalStrength,
		"deviceType":     on.DeviceType,
	})
	return datatypes.JSON(data)
}
