package model

import (
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"time"
)

type (
	DeviceLogReport struct {
		T    int64         `json:"t"`
		Data DeviceLogData `json:"data"`
	}

	DeviceLogData struct {
		LogType   constants.DeviceLogType `json:"logType"`   //上线、下线、属性上报、属性下发等
		MessageId string                  `json:"messageId"` //消息唯一标识符
		Message   string                  `json:"message"`   //消息内容
		Status    string                  `json:"status"`    //状态
	}
)

func NewDeviceLogReport(deviceLogData DeviceLogData) DeviceLogReport {
	return DeviceLogReport{
		T:    time.Now().UnixMilli(),
		Data: deviceLogData,
	}
}
