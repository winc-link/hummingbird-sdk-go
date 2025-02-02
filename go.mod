module github.com/winc-link/hummingbird-sdk-go

go 1.16

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/winc-link/edge-driver-proto v0.0.0-20250202035911-7c16347ede09
	go.uber.org/zap v1.24.0
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
)

require (
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.5.7
	gorm.io/gorm v1.25.12
)
