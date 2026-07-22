package mqtt

import (
	"encoding/base64"
	"fmt"
	"time"
)

// Publisher MQTT 消息发布器
type Publisher struct {
	client *Client
}

// NewPublisher 创建发布器
func NewPublisher(client *Client) *Publisher {
	return &Publisher{client: client}
}

// GetCommand 下发召测命令 (dn/g/<deviceID>)
func (p *Publisher) GetCommand(deviceID string, token int, items []string) error {
	topic := fmt.Sprintf(TopicGetDown, deviceID)
	payload := map[string]interface{}{
		"tk": token,
		"ts": time.Now().Unix(),
		"i":  items,
	}
	return p.client.Publish(topic, payload)
}

// SetCommand 下发参数设置 (dn/s/<deviceID>)
func (p *Publisher) SetCommand(deviceID string, token int, params map[string]interface{}) error {
	topic := fmt.Sprintf(TopicSetDown, deviceID)
	payload := map[string]interface{}{
		"tk": token,
		"ts": time.Now().Unix(),
	}
	for k, v := range params {
		payload[k] = v
	}
	return p.client.Publish(topic, payload)
}

// ActionCommand 下发动作命令 (dn/a/<deviceID>)
func (p *Publisher) ActionCommand(deviceID string, token int, action string, value interface{}) error {
	topic := fmt.Sprintf(TopicActionDown, deviceID)
	payload := map[string]interface{}{
		"tk": token,
		"ts": time.Now().Unix(),
		action: value,
	}
	return p.client.Publish(topic, payload)
}

// Reboot 重启
func (p *Publisher) Reboot(deviceID string, token int) error {
	return p.ActionCommand(deviceID, token, "reboot", nil)
}

// FactoryReset 恢复出厂
func (p *Publisher) FactoryReset(deviceID string, token int) error {
	return p.ActionCommand(deviceID, token, "factory", nil)
}

// SwitchOnOff 开关机
func (p *Publisher) SwitchOnOff(deviceID string, token int, on bool) error {
	val := 0
	if on {
		val = 1
	}
	return p.ActionCommand(deviceID, token, "os", val)
}

// SetActivePower 设置有功功率
func (p *Publisher) SetActivePower(deviceID string, token int, pct int) error {
	return p.ActionCommand(deviceID, token, "ap", pct)
}

// ReportConfirm 回复上报确认帧 (dn/rr/<deviceID>)
func (p *Publisher) ReportConfirm(deviceID string, token int, fields map[string]int) error {
	topic := fmt.Sprintf(TopicReportResp, deviceID)
	payload := map[string]interface{}{
		"tk": token,
		"ts": time.Now().Unix(),
	}
	for k, v := range fields {
		payload[k] = v
	}
	return p.client.Publish(topic, payload)
}

// ReportAck OTA 准备握手 (dn/rr/<deviceID>)
func (p *Publisher) ReportAck(deviceID string) error {
	topic := fmt.Sprintf(TopicReportResp, deviceID)
	payload := map[string]interface{}{
		"tk": 0,
		"ts": time.Now().Unix(),
		"on": 0,
	}
	return p.client.Publish(topic, payload)
}

// OTAStart OTA 升级开始
func (p *Publisher) OTAStart(deviceID string, token int, filename string, fileSize int) error {
	topic := fmt.Sprintf(TopicActionDown, deviceID)
	payload := map[string]interface{}{
		"tk": token,
		"ts": time.Now().Unix(),
		"update": map[string]interface{}{
			"start": map[string]interface{}{
				"fn":  filename,
				"len": fileSize,
			},
		},
	}
	return p.client.Publish(topic, payload)
}

// OTAData OTA 升级数据块
func (p *Publisher) OTAData(deviceID string, token int, offset int, data []byte) error {
	topic := fmt.Sprintf(TopicActionDown, deviceID)
	encoded := base64.StdEncoding.EncodeToString(data)
	payload := map[string]interface{}{
		"tk": token,
		"ts": time.Now().Unix(),
		"update": map[string]interface{}{
			"updating": map[string]interface{}{
				"off": offset,
				"len": len(data),
				"d":   encoded,
			},
		},
	}
	return p.client.Publish(topic, payload)
}

// OTAEnd OTA 升级结束
func (p *Publisher) OTAEnd(deviceID string, token int, md5Str string) error {
	topic := fmt.Sprintf(TopicActionDown, deviceID)
	payload := map[string]interface{}{
		"tk": token,
		"ts": time.Now().Unix(),
		"update": map[string]interface{}{
			"end": map[string]interface{}{
				"md5": md5Str,
			},
		},
	}
	return p.client.Publish(topic, payload)
}
