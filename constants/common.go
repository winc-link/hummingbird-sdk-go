package constants

const (
	TdengineDataDB_PREFIX = "devicedata.device_" //存储设备上报数据
	TdengineLogDB_PREFIX  = "devicelog.log_"     //存储设备日志
)

const (
	PropertyMsg = "property"
	EventMsg    = "event"
)

const (
	EventBusTypePropertyReport = "PROPERTY_REPORT"
	EventBusTypeDeviceStatus   = "DEVICE_STATUS"
	EventBusTypeEventReport    = "EVENT_REPORT"
)

type StorageModel int

const (
	StoreHistory StorageModel = 1 //保存历史
	StoreLatest  StorageModel = 2 //保存最新值
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

	DeviceOnlineEventBus  = "online"
	DeviceOfflineEventBus = "offline"
)

const (
	DataBatchSize     = "data_batch_size"
	DataBatchInterval = "data_batch_interval"
)

const (
	MessageSuccess = "Success"
	MessageError   = "Error"
)

const (
	EventTypeInfo  = "info"
	EventTypeAlert = "alert"
	EventTypeError = "error"
)

type DeviceLogType string

const (
	DeviceLogTypeOnline             DeviceLogType = "设备上线"
	DeviceLogTypeOffline            DeviceLogType = "设备下线"
	DeviceLogTypePropertyReport     DeviceLogType = "属性上报"
	DeviceLogTypeEventReport        DeviceLogType = "事件上报"
	DeviceLogTypePropertyIssue      DeviceLogType = "属性设置"
	DeviceLogTypePropertyIssueReply DeviceLogType = "属性设置响应"
	DeviceLogTypePropertyGet        DeviceLogType = "属性查询"
	DeviceLogTypePropertyGetReply   DeviceLogType = "属性查询响应"
)
