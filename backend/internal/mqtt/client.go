package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTConfig MQTT 配置
type MQTTConfig struct {
	Broker       string `mapstructure:"broker"`
	Port         int    `mapstructure:"port"`
	ClientID     string `mapstructure:"client_id"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	KeepAlive    int    `mapstructure:"keep_alive"`
	CleanSession bool   `mapstructure:"clean_session"`
}

// Client MQTT 客户端
type Client struct {
	conn    mqtt.Client
	handler *MessageHandler
	opts    *mqtt.ClientOptions
}

// MessageHandler 消息处理器
type MessageHandler struct {
	OnReport     func(deviceID string, msg *ReportMessage)
	OnGetResp    func(deviceID string, msg *GetResponse)
	OnSetResp    func(deviceID string, token int, confirm int)
	OnActionResp func(deviceID string, token int, confirm int)
}

// NewClient 创建 MQTT 客户端
func NewClient(cfg MQTTConfig, handler *MessageHandler) *Client {
	broker := fmt.Sprintf("tcp://%s:%d", cfg.Broker, cfg.Port)
	if cfg.Port == 0 {
		broker = cfg.Broker
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(cfg.ClientID).
		SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second).
		SetCleanSession(cfg.CleanSession).
		SetAutoReconnect(true).
		SetConnectRetryInterval(5 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Println("MQTT connected, subscribing...")
			c.Subscribe(TopicSubscribeAll, 1, func(c mqtt.Client, m mqtt.Message) {
				go handler.OnMessage(m)
			})
		})
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	return &Client{
		handler: handler,
		opts:    opts,
	}
}

// Start 启动 MQTT 客户端
func (c *Client) Start() error {
	c.conn = mqtt.NewClient(c.opts)
	token := c.conn.Connect()
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("MQTT connect failed: %w", token.Error())
	}
	log.Println("MQTT client started")
	return nil
}

// Stop 停止 MQTT 客户端
func (c *Client) Stop() {
	if c.conn != nil && c.conn.IsConnected() {
		c.conn.Disconnect(250)
	}
}

// Publish 发布消息（带3秒超时，防止 MQTT 断连时无限阻塞）
func (c *Client) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	token := c.conn.Publish(topic, 1, false, data)
	// 使用 WaitTimeout 而非 Wait，避免 MQTT 断连时无限阻塞
	if !token.WaitTimeout(3 * time.Second) {
		return fmt.Errorf("MQTT publish timeout (topic: %s)", topic)
	}
	return token.Error()
}

// OnMessage 消息分发
func (h *MessageHandler) OnMessage(m mqtt.Message) {
	topic := m.Topic()
	payload := m.Payload()
	deviceID := ExtractDeviceIDFromTopic(topic)
	msgType := ExtractMessageTypeFromTopic(topic)

	if deviceID == "" {
		log.Printf("Cannot extract deviceID from topic: %s", topic)
		return
	}

	log.Printf("[MQTT] %s %s: %s", topic, deviceID, string(payload))

	switch msgType {
	case "r":
		h.handleReport(deviceID, payload)
	case "gr":
		h.handleGetResponse(deviceID, payload)
	case "sr":
		h.handleSetResponse(deviceID, payload)
	case "ar":
		h.handleActionResponse(deviceID, payload)
	default:
		log.Printf("Unknown message type: %s", msgType)
	}
}

func (h *MessageHandler) handleReport(deviceID string, payload []byte) {
	msg, err := ParseReportPayload(payload)
	if err != nil {
		log.Printf("Parse report error: %v", err)
		return
	}
	if h.OnReport != nil {
		h.OnReport(deviceID, msg)
	}
}

func (h *MessageHandler) handleGetResponse(deviceID string, payload []byte) {
	msg, err := ParseGetResponsePayload(payload)
	if err != nil {
		log.Printf("Parse get response error: %v", err)
		return
	}
	if h.OnGetResp != nil {
		h.OnGetResp(deviceID, msg)
	}
}

func (h *MessageHandler) handleSetResponse(deviceID string, payload []byte) {
	var msg ResponseMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("Parse set response error: %v", err)
		return
	}
	if h.OnSetResp != nil {
		h.OnSetResp(deviceID, msg.Token, 0)
	}
}

func (h *MessageHandler) handleActionResponse(deviceID string, payload []byte) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		log.Printf("Parse action response error: %v", err)
		return
	}

	token := 0
	if v, ok := raw["tk"]; ok {
		switch t := v.(type) {
		case float64:
			token = int(t)
		case int:
			token = t
		}
	}

	confirm := 0
	if v, ok := raw["update"]; ok {
		switch t := v.(type) {
		case float64:
			confirm = int(t)
		case int:
			confirm = t
		}
	}

	if h.OnActionResp != nil {
		h.OnActionResp(deviceID, token, confirm)
	}
}
