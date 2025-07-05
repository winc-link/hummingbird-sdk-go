package main

import (
	"context"
	"fmt"
	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"github.com/winc-link/hummingbird-sdk-go/datadb/tdengine"
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
			Host:              "124.223.78.197",
			Port:              58090,
			Type:              "mqtt",
			MessageQueueTopic: "eventbus/in",
		}),
		service.WithCustomDataBasesConfig(&service.DataBasesConnConfig{
			Type: constants.DataBasesTdengine,
			Tdengine: tdengine.DbClient{
				Dsn: "root:c1Fps2rDzdudIiqQ@ws(124.223.78.197:6041)/hummingbird",
			},
		}), service.WithCustomMetaBasesConfig(&service.MetaBasesConnConfig{
			Type: constants.MetadataMysql,
			Dns:  "hummingbird:fK1hrMEObgRfQMie@tcp(124.223.78.197:3306)/hummingbird?charset=utf8mb4&parseTime=True&loc=Local&timeout=2s",
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

func MsgEventReport(driverService *service.DriverService) {
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.EventReport("41759677", model.EventReport{
			CommonRequest: model.CommonRequest{
				MsgId: "132",
				Time:  time.Now().Unix(),
			},
			Data: map[string]map[string]interface{}{
				"Err": map[string]interface{}{
					"Code": "400",
				},
			},
		})
		fmt.Println(resp)
	}
}

func MsgReport(driverService *service.DriverService) {
	for {
		time.Sleep(1 * time.Second)
		resp, _ := driverService.PropertyReport("47624160", model.NewPropertyReport(model.NewDefaultCommonRequest(), map[string]interface{}{
			"HK_RealTimeData@0x0203": 9,
			"HK_RealTimeData@0x0204": "abc",
			"HK_RealTimeData@0x0205": true,
		}))
		fmt.Println(resp)
		//value++
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
