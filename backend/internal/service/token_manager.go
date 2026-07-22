package service

import (
	"log"
	"sync"
	"time"

	"iot-platform/internal/mqtt"
)

// MQTTResponse MQTT 响应
type MQTTResponse struct {
	Token   int
	Confirm int
	Data    *mqtt.GetResponse
}

// pendingRequest 待响应的请求
type pendingRequest struct {
	deviceID string
	token    uint16
	ch       chan MQTTResponse
	created  time.Time
}

// TokenManager MQTT 请求-响应 token 管理器
type TokenManager struct {
	mu       sync.RWMutex
	pending  map[string]*pendingRequest // key: deviceID:token
	counter  uint16
}

// NewTokenManager 创建 token 管理器
func NewTokenManager() *TokenManager {
	tm := &TokenManager{
		pending: make(map[string]*pendingRequest),
	}
	// 启动超时清理协程
	go tm.cleanupLoop()
	return tm
}

func (tm *TokenManager) key(deviceID string, token uint16) string {
	return deviceID + ":" + string(rune(token))
}

// NextToken 生成下一个 token
func (tm *TokenManager) NextToken() uint16 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.counter++
	if tm.counter == 0 {
		tm.counter = 1
	}
	return tm.counter
}

// Register 注册一个待响应的请求，返回响应 channel
func (tm *TokenManager) Register(deviceID string, token uint16) chan MQTTResponse {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	ch := make(chan MQTTResponse, 1)
	req := &pendingRequest{
		deviceID: deviceID,
		token:    token,
		ch:       ch,
		created:  time.Now(),
	}
	tm.mu.Lock()
	tm.pending[tm.key(deviceID, token)] = req
	tm.mu.Unlock()

	return ch
}

// Resolve 解析响应，将结果发送到等待的 channel
func (tm *TokenManager) Resolve(deviceID string, token uint16, resp MQTTResponse) {
	tm.mu.Lock()
	req, ok := tm.pending[tm.key(deviceID, token)]
	if ok {
		delete(tm.pending, tm.key(deviceID, token))
	}
	tm.mu.Unlock()

	if ok {
		select {
		case req.ch <- resp:
		default:
		}
	}
}

// RequestAndWait 发送请求并等待响应（带超时）
func (tm *TokenManager) RequestAndWait(deviceID string, token uint16, timeout time.Duration) (*MQTTResponse, bool) {
	ch := tm.Register(deviceID, token)
	select {
	case resp := <-ch:
		return &resp, true
	case <-time.After(timeout):
		tm.Cancel(deviceID, token)
		return nil, false
	}
}

// Cancel 取消待响应的请求
func (tm *TokenManager) Cancel(deviceID string, token uint16) {
	tm.mu.Lock()
	delete(tm.pending, tm.key(deviceID, token))
	tm.mu.Unlock()
}

// cleanupLoop 定期清理超时的请求
func (tm *TokenManager) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		tm.mu.Lock()
		now := time.Now()
		for k, req := range tm.pending {
			if now.Sub(req.created) > 60*time.Second {
				log.Printf("TokenManager: cleanup expired request %s:%d", req.deviceID, req.token)
				delete(tm.pending, k)
			}
		}
		tm.mu.Unlock()
	}
}
