package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"github.com/winc-link/hummingbird-sdk-go/datadb/influxdb"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"github.com/winc-link/hummingbird-sdk-go/service"
	"math/rand/v2"
	"time"
)

// go run cmd/main.go -c cmd/res/configuration.toml
func main() {
	fmt.Println("test....")

	driverService := service.NewDriverService("test",
		service.WithCustomRedisBasesConfig(&service.RedisBasesConnConfig{
			Address:  "124.221.36.14:6379",
			Password: "",
			DB:       0,
		}),
		//{hummingbird device-data http://124.221.36.14:8086 6S4LFh_kP0-RHqIYjYlXgvGfXOgMIkqMkinZDePKiXbcmgIQzQcm5mV5GfSQEDVqcKzQ5WIixO7AEmwHQ17JmQ==
		service.WithCustomDataBasesConfig(&service.DataBasesConnConfig{
			Type: constants.DataBasesInfluxdb,
			InfluxDB: influxdb.DbClient{
				Org:       "hummingbird",
				Bucket:    "device-data",
				LogBucket: "device-log",
				Url:       "http://124.221.36.14:8086",
				Token:     "6S4LFh_kP0-RHqIYjYlXgvGfXOgMIkqMkinZDePKiXbcmgIQzQcm5mV5GfSQEDVqcKzQ5WIixO7AEmwHQ17JmQ==",
			},

			//Type: constants.DataBasesTdengine,
			//Tdengine: tdengine.DbClient{
			//	Dsn: "root:taosdata@ws(127.0.0.1:6041)/devicedata",
			//},
		}), service.WithCustomMetaBasesConfig(&service.MetaBasesConnConfig{
			Type: constants.MetadataMysql,
			Dns:  "root:!@#12345678.@tcp(124.221.36.14:3306)/hummingbird?charset=utf8mb4&parseTime=True&loc=Local&timeout=2s",
		}))

	//go func() {
	//	MsgReport(driverService)
	//}()

	go func() {
		MsgReport22(driverService)
	}()
	//
	//go func() {
	//	MsgReport2(driverService)
	//}()
	//go func() {
	//	MsgReport3(driverService)
	//}()

	//go func() {
	//	MsgEventReport(driverService)
	//}()
	//
	//go func() {
	//	MsgEventReport(driverService)
	//}()

	//go func() {
	//	DeviceLogReport(driverService)
	//}()

	var d driverTest
	err := driverService.Start(d)
	if err != nil {
		panic(err)
	}

}

func MsgEventReport(driverService *service.DriverService) {
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.EventReport("72816757", model.EventReport{
			//CommonRequest: model.CommonRequest{
			//	MsgId: "132",
			//	Time:  time.Now().Unix(),
			//},
			Data: map[string]map[string]interface{}{
				"Error": map[string]interface{}{
					"param1": 400,
					"param":  false,
				},
			},
		})
		fmt.Println(resp)
	}
}

func MsgReport(driverService *service.DriverService) {
	temp := 100
	hum := 1
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.PropertyReport("72816757", model.NewPropertyReport(model.NewDefaultCommonRequest(), map[string]interface{}{
			"battery_percentage": Rand50To60(),
			"pir_state":          0,
			"temp":               Rand30To40(),
			"co_value":           Rand70To80(),
			//"hum":                Rand50To60(),
			//"switch": true,
		}))
		temp++
		hum++
		fmt.Println(resp)
	}
}

func MsgReport22(driverService *service.DriverService) {
	temp := 100
	hum := 1
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.PropertyReport("37159496", model.NewPropertyReport(model.NewDefaultCommonRequest(), map[string]interface{}{
			"battery_percentage": Rand50To60(),
			"pir_state":          0,
			"temp":               Rand70To80(),
			"co_value":           Rand30To40(),
			//"hum":                Rand50To60(),
			//"switch": true,
		}))
		temp++
		hum++
		fmt.Println(resp)
	}
}

func Rand30To40() int {
	return rand.IntN(11) + 30 // 0~10 再 +90 → 90~100
}

func Rand50To60() int {
	return rand.IntN(11) + 50 // 0~10 再 +90 → 90~100
}

func Rand70To80() int {
	return rand.IntN(11) + 70 // 0~10 再 +90 → 90~100
}

func MsgReport2(driverService *service.DriverService) {
	temp := 300
	hum := 100
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.PropertyReport("083228610", model.NewPropertyReport(model.NewDefaultCommonRequest(), map[string]interface{}{
			"temp":   temp,
			"hum":    hum,
			"switch": false,
		}))
		temp++
		hum++
		fmt.Println(resp)
	}
}

func MsgReport3(driverService *service.DriverService) {
	temp := 300
	hum := 100
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.PropertyReport("08322862", model.NewPropertyReport(model.NewDefaultCommonRequest(), map[string]interface{}{
			"temp":   temp,
			"hum":    hum,
			"switch": false,
		}))
		temp++
		hum++
		fmt.Println(resp)
	}
}
func DeviceLogReport(driverService *service.DriverService) {
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.DeviceLogReport("123", model.NewDeviceLogReport(model.DeviceLogData{
			LogType:   constants.DeviceLogTypeOnline,
			MessageId: uuid.New().String(),
			Message:   "记录包含多字节字符在内的字符串，如中文字符。每个 NCHAR 字符占用 4 字节的存储空间。字符串两端使用单引号引用，字符串内的单引号需用转义字符 。NCHAR 使用时须指定字符串大小，类型为 NCHAR(10) 的列表示此列的字符串最多存储 10 个 NCHAR 字符。如果用户字符串长度超出声明长度，将会报错。",
			Status:    constants.MessageSuccess,
		}))
		fmt.Println(resp)
	}
}

type driverTest struct {
}

func (d driverTest) HandlePropertyReportDebug(ctx context.Context, deviceId string, data model.PropertyReport) error {
	//TODO implement me
	panic("implement me")
}

func (d driverTest) HandleEventReportDebug(ctx context.Context, deviceId string, data model.EventReport) error {
	//TODO implement me
	panic("implement me")
}

func (d driverTest) DeviceNotify(ctx context.Context, t commons.DeviceNotifyType, deviceId string, device model.Device) error {
	//TODO implement me
	panic("implement me")
}

func (d driverTest) ProductNotify(ctx context.Context, t commons.ProductNotifyType, productId string, product model.Product) error {
	//TODO implement me
	panic("implement me")
}

func (d driverTest) Stop(ctx context.Context) error {
	//TODO implement me
	return nil
}

func (d driverTest) HandlePropertySet(ctx context.Context, deviceId string, data model.PropertySet) error {
	//TODO implement me
	panic("implement me")
}

func (d driverTest) HandlePropertyGet(ctx context.Context, deviceId string, data model.PropertyGet) error {
	//TODO implement me
	panic("implement me")
}

func (d driverTest) HandleServiceExecute(ctx context.Context, deviceId string, data model.ServiceExecuteRequest) error {
	//TODO implement me
	panic("implement me")
}
