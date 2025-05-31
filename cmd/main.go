package main

import (
	"context"
	"fmt"
	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"github.com/winc-link/hummingbird-sdk-go/datadb/influxdb"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"github.com/winc-link/hummingbird-sdk-go/service"
	"time"
)

// go run cmd/main.go -c cmd/res/configuration.toml
func main() {
	fmt.Println("test....")

	driverService := service.NewDriverService("test",
		service.WithCustomMessageQueueConfig(&service.MessageQueueConnConfig{
			Protocol:          "tcp",
			Host:              "*.221.189.137",
			Port:              58090,
			Type:              "mqtt",
			MessageQueueTopic: "eventbus/in",
		}),
		service.WithCustomDataBasesConfig(&service.DataBasesConnConfig{
			Type: constants.DataBasesInfluxdb,
			InfluxDB: influxdb.DbClient{
				Org:    "hummingbird",
				Bucket: "hummingbird",
				Url:    "http://*.221.189.137:8086",
				Token:  "6S4LFh_kP0-RHqIYjYlXgvGfXOgMIkqMkinZDePKiXbcmgIQzQcm5mV5GfSQEDVqcKzQ5WIixO7AEmwHQ17JmQ==",
			},
		}), service.WithCustomMetaBasesConfig(&service.MetaBasesConnConfig{
			Type: constants.MetadataMysql,
			Dns:  "root:!@#12345678.@tcp(*.221.189.137:3306)/hummingbird?charset=utf8mb4&parseTime=True&loc=Local&timeout=2s",
		}))

	go func() {
		MsgReport(driverService)
	}()
	var d driverTest
	err := driverService.Start(d)
	if err != nil {
		panic(err)
	}

}

func MsgReport(driverService *service.DriverService) {
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.PropertyReport("50216271", model.NewPropertyReport(model.NewDefaultCommonRequest(), map[string]interface{}{
			"abc": "123",
			"efg": "456",
		}))
		fmt.Println(resp)
	}
}

type driverTest struct {
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
