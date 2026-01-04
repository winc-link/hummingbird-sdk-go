package model

type (
	BatchInsertDeviceLogData struct {
		T        int64         `json:"t"`
		DeviceID string        `json:"deviceId"`
		Data     DeviceLogData `json:"data"`
	}

	BatchInsertPropertyData struct {
		DeviceID string                 `json:"deviceId"`
		T        int64                  `json:"t"`
		Data     map[string]interface{} `json:"data"`
	}

	BatchInsertEventData struct {
		DeviceID  string                 `json:"deviceId"`
		T         int64                  `json:"t"`
		EventType string                 `json:"eventType"`
		Data      map[string]interface{} `json:"data"`
	}
)
