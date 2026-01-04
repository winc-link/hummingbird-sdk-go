package config

const (
	MetaBasesType_Mysql  int32 = 0
	MetaBasesType_Sqlite int32 = 1
)

type HummingbirdConfig struct {
	MetaBases  MetaBases  `json:"meta_bases"`  //数据库配置
	DataBases  DataBases  `json:"data_bases"`  //时序数据库配置
	RedisBases RedisBases `json:"redis_bases"` //redis 配置
}

type MetaBases struct {
	Type   int32  `json:"type"`   //mysql or sqlite
	Source string `json:"source"` //连接信息
}

type DataBases struct {
	Type       string          `json:"type"` //influxdb or tdengine or clickhouse
	Tdengine   *TDengineSource `json:"tdengine"`
	InfluxDB   *InfluxDB       `json:"influxdb"`
	ClickHouse *ClickHouse     `json:"clickhouse"`
}

type ClickHouse struct {
	Addr     []string `json:"addr"`
	Database string   `json:"database"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}

type InfluxDB struct {
	Org    string `json:"org"`
	Bucket string `json:"bucket"`
	Url    string `json:"url"`
	Token  string `json:"token"`
}

type TDengineSource struct {
	Dsn string `json:"dsn"`
}

type RedisBases struct {
	Address  string `json:"address"`
	Password string `json:"password"`
	DB       int32  `json:"db"`
}
