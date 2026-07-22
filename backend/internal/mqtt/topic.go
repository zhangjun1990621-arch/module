package mqtt

const (
	// 上行 Topic 模板 (设备 → 平台)
	TopicReportUp   = "up/r/%s"  // 设备上报
	TopicGetResp    = "up/gr/%s" // get 响应
	TopicSetResp    = "up/sr/%s" // set 响应
	TopicActionResp = "up/ar/%s" // action 响应

	// 下行 Topic 模板 (平台 → 设备)
	TopicGetDown    = "dn/g/%s"  // get 请求
	TopicSetDown    = "dn/s/%s"  // set 请求
	TopicActionDown = "dn/a/%s"  // action 请求
	TopicReportResp = "dn/rr/%s" // report 响应确认

	// 订阅 Topic
	TopicSubscribeAll = "up/+/#"
)
