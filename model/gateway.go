package model

type (
	// GatewayControlSet 设备向云端上报事件
	GatewayControlSet struct {
		ControlType int    `json:"control_type"`
		Data        string `json:"data"`
	}
)
