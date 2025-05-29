package main

import (
	"context"
	"fmt"
	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"github.com/winc-link/hummingbird-sdk-go/service"
	"time"
)

// go run cmd/main.go -c cmd/res/configuration.toml
func main() {
	fmt.Println("test....")

	driverService := service.NewDriverService("test")

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
		resp, _ := driverService.PropertyReport("123", model.NewPropertyReport(model.NewDefaultCommonRequest(), map[string]interface{}{
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
