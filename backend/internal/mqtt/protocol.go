package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// 短键 → 全名映射
var ShortKeyMap = map[string]string{
	"tk":    "token",
	"ts":    "timestamp",
	"cs":    "signalStrength",
	"sw":    "software",
	"hw":    "hardware",
	"dt":    "deviceType",
	"md":    "model",
	"sn":    "serialNo",
	"cp":    "capacity",
	"fw":    "firmware",
	"dc":    "dc",
	"ac":    "ac",
	"eov":   "overVoltage",
	"eov_r": "overVoltageRecover",
	"euv":   "underVoltage",
	"euv_r": "underVoltageRecover",
	"elc":   "localControl",
	"lc":    "localCtrl",
	"ap":    "activePowerPct",
	"rp":    "reactivePowerPct",
	"rc":    "remoteCtrl",
	"os":    "onOffSwitch",
	"fd":    "frozenData",
}

type ReportMessage struct {
	Token     int            `json:"tk"`
	Timestamp int64          `json:"ts"`
	On        *OnlineEvent   `json:"on,omitempty"`
	Off       interface{}    `json:"off,omitempty"`
	HB        *Heartbeat     `json:"hb,omitempty"`
	DC        *DCData        `json:"dc,omitempty"`
	AC        *ACData        `json:"ac,omitempty"`
	EOV       *EventDetail   `json:"eov,omitempty"`
	EOVR      *EventDetail   `json:"eov_r,omitempty"`
	EUV       *EventDetail   `json:"euv,omitempty"`
	EUVR      *EventDetail   `json:"euv_r,omitempty"`
	ELC       *LocalControl  `json:"elc,omitempty"`
}

type OnlineEvent struct {
	SignalStrength int    `json:"cs"`
	Software       string `json:"sw"`
	Hardware       string `json:"hw"`
	DeviceType     string `json:"dt"`
}

type Heartbeat struct {
	SignalStrength int `json:"cs"`
}

type DCData struct {
	Paths    []string  `json:"paths,omitempty"`
	Voltages []float64 `json:"v,omitempty"`
	Currents []float64 `json:"c,omitempty"`
	Powers   []int     `json:"p,omitempty"`
}

type ACData struct {
	Phases        int       `json:"ph,omitempty"`
	Voltages      []float64 `json:"v,omitempty"`
	Currents      []float64 `json:"c,omitempty"`
	ActivePower   int       `json:"p,omitempty"`
	ReactivePower int       `json:"q,omitempty"`
	PowerFactor   float64   `json:"pf,omitempty"`
	Frequency     float64   `json:"f,omitempty"`
}

type EventDetail struct {
	Data      map[string]interface{} `json:"-"`
	Timestamp int64                  `json:"ts,omitempty"`
}

func (e *EventDetail) UnmarshalJSON(data []byte) error {
	e.Data = make(map[string]interface{})
	return json.Unmarshal(data, &e.Data)
}

type LocalControl struct {
	Cycle       int     `json:"cy,omitempty"`
	StartTime   int     `json:"st,omitempty"`
	EndTime     int     `json:"en,omitempty"`
	VoltageHigh float64 `json:"vh,omitempty"`
	VoltageLow  float64 `json:"vl,omitempty"`
}

type GetResponse struct {
	Token      int           `json:"tk"`
	Timestamp  int64         `json:"ts"`
	Software   string        `json:"sw,omitempty"`
	Hardware   string        `json:"hw,omitempty"`
	Model      string        `json:"md,omitempty"`
	SerialNo   string        `json:"sn,omitempty"`
	Capacity   int           `json:"cp,omitempty"`
	Firmware   string        `json:"fw,omitempty"`
	DeviceType string        `json:"dt,omitempty"`
	DC         *DCData       `json:"dc,omitempty"`
	AC         *ACData       `json:"ac,omitempty"`
}

type ResponseMessage struct {
	Token     int         `json:"tk"`
	Timestamp int64       `json:"ts"`
	Data      interface{} `json:"-"`
}

func ParseReportPayload(payload []byte) (*ReportMessage, error) {
	var msg ReportMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, fmt.Errorf("parse report payload: %w", err)
	}
	return &msg, nil
}

func ParseGetResponsePayload(payload []byte) (*GetResponse, error) {
	var msg GetResponse
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, fmt.Errorf("parse get response payload: %w", err)
	}
	return &msg, nil
}

func ExtractDeviceIDFromTopic(topic string) string {
	parts := strings.Split(topic, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func ExtractMessageTypeFromTopic(topic string) string {
	parts := strings.Split(topic, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func LogPayload(prefix string, payload []byte) {
	log.Printf("%s %s", prefix, string(payload))
}
