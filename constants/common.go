package constants

const (
	DB_PREFIX = "device_"
)

const (
	EventBusTypePropertyReport = "PROPERTY_REPORT"
	EventBusTypeDeviceStatus   = "DEVICE_STATUS"
	EventBusTypeEventReport    = "EVENT_REPORT"
)

const (
	MessageQueueMqtt = "mqtt"
)

type MetadataType string

const (
	MetadataMysql  MetadataType = "mysql"
	MetadataSqlite MetadataType = "sqlite"
)

type DataBasesType string

const (
	DataBasesInfluxdb   DataBasesType = "influxdb"
	DataBasesTdengine   DataBasesType = "tdengine"
	DataBasesClickhouse DataBasesType = "clickhouse"
)

const (
	DeviceOnline  = "在线"
	DeviceOffline = "离线"
)
