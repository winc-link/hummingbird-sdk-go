/*******************************************************************************
 * Copyright 2017.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package service

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/interfaces"
	"github.com/winc-link/hummingbird-sdk-go/internal/cache"
	"github.com/winc-link/hummingbird-sdk-go/internal/client"
	"github.com/winc-link/hummingbird-sdk-go/internal/config"
	"github.com/winc-link/hummingbird-sdk-go/internal/logger"
	"github.com/winc-link/hummingbird-sdk-go/internal/server"
	"github.com/winc-link/hummingbird-sdk-go/internal/snowflake"
	"github.com/winc-link/hummingbird-sdk-go/model"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/winc-link/edge-driver-proto/cloudinstance"
	"github.com/winc-link/edge-driver-proto/drivercommon"
	"github.com/winc-link/edge-driver-proto/driverdevice"
	"google.golang.org/grpc/status"
)

type DriverService struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	//platform          commons.IotPlatform
	cfg               *config.DriverConfig
	driverServiceName string
	logger            logger.Logger
	deviceCache       cache.DeviceProvider
	productCache      cache.ProductProvider
	driver            interfaces.Driver
	rpcClient         *client.ResourceClient
	rpcServer         *server.RpcService
	mqttClient        mqtt.Client
	baseMessage       commons.BaseMessage
	node              *snowflake.Worker
	dbClient          *gorm.DB
	readyChan         chan struct{}
}

func NewDriverService(serviceName string) *DriverService {
	var (
		wg         sync.WaitGroup
		err        error
		cfg        *config.DriverConfig
		log        logger.Logger
		coreClient *client.ResourceClient
		node       *snowflake.Worker
		db         *gorm.DB
	)

	flag.StringVar(&config.FilePath, "c", config.DefaultConfigFilePath, "./driver -c configFile")
	flag.Parse()
	if cfg, err = config.ParseConfig(); err != nil {
		fmt.Println(err)

		os.Exit(-1)
	}
	if err = cfg.ValidateConfig(); err != nil {
		os.Exit(-1)
	}

	log = logger.NewLogger(cfg.Logger.FileName, cfg.Logger.LogLevel, serviceName)

	// Start rpc client
	if coreClient, err = client.NewCoreClient(cfg.Clients[config.Core]); err != nil {
		log.Errorf("new resource client error: %v rpcServer", err)
		os.Exit(-1)
	}
	// Snowflake node
	if node, err = snowflake.NewWorker(1); err != nil {
		log.Errorf("new msg id generator error: %v rpcServer", err)
		os.Exit(-1)
	}

	hummingbirdConfig, err := coreClient.GetHummingbirdConfig(context.Background(), nil)
	if err != nil {
		log.Errorf("get hummingbird config error: %v", err)
		os.Exit(-1)
	}
	switch hummingbirdConfig.MetaBases.Type {
	case drivercommon.MetaBasesType_Mysql:
		dsn := hummingbirdConfig.MetaBases.Source
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Errorf("open db error: %v rpcServer", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	driverService := &DriverService{
		ctx:               ctx,
		cancel:            cancel,
		wg:                &wg,
		rpcClient:         coreClient,
		logger:            log,
		cfg:               cfg,
		node:              node,
		driverServiceName: serviceName,
		dbClient:          db,
	}
	if err = driverService.buildRpcBaseMessage(); err != nil {
		log.Error("buildRpcBaseMessage error:", err)
		os.Exit(-1)
	}

	if err = driverService.reportDriverInfo(); err != nil {
		log.Error("reportDriverInfo error:", err)
		os.Exit(-1)
	}

	if err = driverService.initCache(); err != nil {
		log.Error("initCache error:", err)
		os.Exit(-1)
	}

	return driverService
}

func (d *DriverService) buildRpcBaseMessage() error {
	var baseMessage commons.BaseMessage
	baseMessage.DriverInstanceId = d.cfg.GetServiceID()
	d.baseMessage = baseMessage
	return nil
}

func (d *DriverService) initCache() error {
	// Sync device
	if deviceCache, err := cache.InitDeviceCache(d.baseMessage, d.rpcClient, d.logger); err != nil {
		d.logger.Errorf("sync device error: %rpcServer", err)
		os.Exit(-1)
	} else {
		d.deviceCache = deviceCache
	}

	// Sync product
	if productCache, err := cache.InitProductCache(d.baseMessage, d.rpcClient, d.logger); err != nil {
		d.logger.Errorf("sync tsl error: %s rpcServer", err)
		os.Exit(-1)
	} else {
		d.productCache = productCache
	}
	return nil
}

func (d *DriverService) reportDriverInfo() error {
	// 上报驱动信息
	timeoutContext, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)

	defer cancelFunc()
	var reportPlatformInfoRequest cloudinstance.DriverReportPlatformInfoRequest
	reportPlatformInfoRequest.DriverInstanceId = d.cfg.GetServiceID()
	driverReportPlatformResp, err := d.rpcClient.CloudInstanceServiceClient.DriverReportPlatformInfo(timeoutContext, &reportPlatformInfoRequest)
	if err != nil {
		os.Exit(-1)
	}
	if !driverReportPlatformResp.BaseResponse.Success {
		return errors.New(driverReportPlatformResp.BaseResponse.ErrorMessage)
	}
	return nil
}

func (d *DriverService) start(driver interfaces.Driver) error {
	var err error
	if driver == nil {
		return errors.New("driver unimplemented")
	}
	d.driver = driver

	// rpc server
	d.rpcServer, err = server.NewRpcService(d.ctx, d.wg, d.cancel, d.cfg.Service.Server, d.deviceCache, d.productCache,
		d.driver, d.rpcClient, d.logger)
	if err != nil {
		os.Exit(-1)
	}
	go d.waitSignalsExit()
	_ = d.rpcServer.Start()

	return nil
}

func (d *DriverService) waitSignalsExit() {
	stopSignalCh := make(chan os.Signal, 1)
	signal.Notify(stopSignalCh, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, os.Interrupt)

	for {
		select {
		case <-stopSignalCh:
			d.logger.Info("got stop signal, exit...")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			if err := d.driver.Stop(ctx); err != nil {
				d.logger.Errorf("call protocol driver stop function error: %s", err)
			}
			cancel()
			d.cancel()
			return
		case <-d.ctx.Done():
			d.logger.Info("inner cancel executed, exit...")
			d.wg.Add(1)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			if err := d.driver.Stop(ctx); err != nil {
				d.logger.Errorf("call protocol driver stop function error: %s", err)
			}
			cancel()
			d.wg.Done()
			return
		}
	}
}

func (d *DriverService) propertySetResponse(cid string, data model.PropertySetResponse) error {
	msg, err := commons.TransformToProtoMsg(cid, commons.PropertySetResponse, data, d.baseMessage)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return errors.New(status.Convert(err).Message())
	}
	return nil
}

func (d *DriverService) propertyGetResponse(cid string, data model.PropertyGetResponse) error {
	msg, err := commons.TransformToProtoMsg(cid, commons.PropertyGetResponse, data, d.baseMessage)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return errors.New(status.Convert(err).Message())
	}
	return nil
}

func (d *DriverService) serviceExecuteResponse(cid string, data model.ServiceExecuteResponse) error {
	msg, err := commons.TransformToProtoMsg(cid, commons.ServiceExecuteResponse, data, d.baseMessage)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return errors.New(status.Convert(err).Message())
	}
	return nil
}

func (d *DriverService) propertyReport(cid string, data model.PropertyReport) (model.CommonResponse, error) {
	msgId := d.node.GetId().String()
	data.MsgId = msgId
	msg, err := commons.TransformToProtoMsg(cid, commons.PropertyReport, data, d.baseMessage)
	if err != nil {
		return model.CommonResponse{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	thingModelResp := new(drivercommon.CommonResponse)
	if thingModelResp, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return model.CommonResponse{}, errors.New(status.Convert(err).Message())
	}

	return model.NewCommonResponse(thingModelResp), nil
}

func (d *DriverService) eventReport(cid string, data model.EventReport) (model.CommonResponse, error) {
	msgId := d.node.GetId().String()
	data.MsgId = msgId
	msg, err := commons.TransformToProtoMsg(cid, commons.EventReport, data, d.baseMessage)
	if err != nil {
		return model.CommonResponse{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	thingModelResp := new(drivercommon.CommonResponse)
	if thingModelResp, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return model.CommonResponse{}, errors.New(status.Convert(err).Message())
	}

	return model.NewCommonResponse(thingModelResp), nil
}

func (d *DriverService) batchReport(cid string, data model.BatchReport) (model.CommonResponse, error) {
	msgId := d.node.GetId().String()
	data.MsgId = msgId
	msg, err := commons.TransformToProtoMsg(cid, commons.BatchReport, data, d.baseMessage)
	if err != nil {
		return model.CommonResponse{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	thingModelResp := new(drivercommon.CommonResponse)
	if thingModelResp, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return model.CommonResponse{}, errors.New(status.Convert(err).Message())
	}

	return model.NewCommonResponse(thingModelResp), nil
}

func (d *DriverService) propertyDesiredGet(deviceId string, data model.PropertyDesiredGet) (model.PropertyDesiredGetResponse, error) {
	msgId := d.node.GetId().String()
	data.MsgId = msgId
	msg, err := commons.TransformToProtoMsg(deviceId, commons.PropertyDesiredGet, data, d.baseMessage)
	if err != nil {
		return model.PropertyDesiredGetResponse{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	thingModelResp := new(drivercommon.CommonResponse)
	if thingModelResp, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return model.PropertyDesiredGetResponse{}, errors.New(status.Convert(err).Message())
	}
	d.logger.Info(thingModelResp)
	return model.PropertyDesiredGetResponse{}, nil
}

func (d *DriverService) propertyDesiredDelete(deviceId string, data model.PropertyDesiredDelete) (model.CommonResponse, error) {
	msgId := d.node.GetId().String()
	data.MsgId = msgId
	msg, err := commons.TransformToProtoMsg(deviceId, commons.PropertyDesiredDelete, data, d.baseMessage)
	if err != nil {
		return model.CommonResponse{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	thingModelResp := new(drivercommon.CommonResponse)
	if thingModelResp, err = d.rpcClient.ThingModelMsgReport(ctx, msg); err != nil {
		return model.CommonResponse{}, errors.New(status.Convert(err).Message())
	}

	return model.NewCommonResponse(thingModelResp), nil
}

func (d *DriverService) connectIotPlatform(deviceId string) error {
	var (
		err  error
		resp *driverdevice.ConnectIotPlatformResponse
	)
	if len(deviceId) == 0 {
		return errors.New("required device id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req := driverdevice.ConnectIotPlatformRequest{
		BaseRequest: d.baseMessage.BuildBaseRequest(),
		DeviceId:    deviceId,
	}
	if resp, err = d.rpcClient.ConnectIotPlatform(ctx, &req); err != nil {
		return errors.New(status.Convert(err).Message())
	}
	if resp != nil {
		if !resp.BaseResponse.Success {
			return errors.New(resp.BaseResponse.ErrorMessage)
		}
		if resp.Data.Status == driverdevice.ConnectStatus_ONLINE {
			device, ok := d.deviceCache.SearchById(deviceId)
			if ok {
				device.Status = commons.DeviceOnline
				d.deviceCache.Update(device)
			}
			return nil
		}
	}
	return errors.New("unKnow error")
}

func (d *DriverService) disconnectIotPlatform(deviceId string) error {
	var (
		err  error
		resp *driverdevice.DisconnectIotPlatformResponse
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req := driverdevice.DisconnectIotPlatformRequest{
		BaseRequest: d.baseMessage.BuildBaseRequest(),
		DeviceId:    deviceId,
	}
	if resp, err = d.rpcClient.DisconnectIotPlatform(ctx, &req); err != nil {
		return errors.New(status.Convert(err).Message())
	}
	if resp != nil {
		if !resp.BaseResponse.Success {
			return errors.New(resp.BaseResponse.ErrorMessage)
		}
		if resp.Data.Status == driverdevice.ConnectStatus_ONLINE {

		} else if resp.Data.Status == driverdevice.ConnectStatus_OFFLINE {
			device, ok := d.deviceCache.SearchById(deviceId)
			if ok {
				device.Status = commons.DeviceOffline
				d.deviceCache.Update(device)
			}
			return nil
		}
	}
	return errors.New("unKnow error")
}

func (d *DriverService) getConnectStatus(deviceId string) (commons.DeviceConnectStatus, error) {
	var (
		err  error
		resp *driverdevice.GetDeviceConnectStatusResponse
	)

	if len(deviceId) == 0 {
		return "", errors.New("required device cid")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req := driverdevice.GetDeviceConnectStatusRequest{
		BaseRequest: d.baseMessage.BuildBaseRequest(),
		DeviceId:    deviceId,
	}
	if resp, err = d.rpcClient.GetDeviceConnectStatus(ctx, &req); err != nil {
		return "", errors.New(status.Convert(err).Message())
	}
	if resp != nil {
		if !resp.BaseResponse.Success {
			return "", errors.New(resp.BaseResponse.ErrorMessage)
		}
		if resp.Data.Status == driverdevice.ConnectStatus_ONLINE {
			return commons.Online, nil
		} else if resp.Data.Status == driverdevice.ConnectStatus_OFFLINE {
			return commons.Offline, nil
		}
	}
	return "", errors.New("unKnow error")
}

func (d *DriverService) getDeviceList() []model.Device {
	var devices []model.Device
	for _, v := range d.deviceCache.All() {
		devices = append(devices, v)
	}
	return devices
}

func (d *DriverService) getDeviceById(deviceId string) (model.Device, bool) {
	device, ok := d.deviceCache.SearchById(deviceId)
	if !ok {
		return model.Device{}, false
	}
	return device, true
}

func (d *DriverService) createDevice(addDevice model.AddDevice) (device model.Device, err error) {
	var (
		resp *driverdevice.CreateDeviceRequestResponse
	)

	if addDevice.ProductId == "" || addDevice.Name == "" || addDevice.DeviceSn == "" {
		err = errors.New("param failed")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	reqDevice := new(driverdevice.AddDevice)
	reqDevice.Name = addDevice.Name
	reqDevice.ProductId = addDevice.ProductId
	reqDevice.DeviceSn = addDevice.DeviceSn
	reqDevice.Description = addDevice.Description
	reqDevice.External = addDevice.External
	req := driverdevice.CreateDeviceRequest{
		BaseRequest: d.baseMessage.BuildBaseRequest(),
		Device:      reqDevice,
	}
	if resp, err = d.rpcClient.CreateDevice(ctx, &req); err != nil {
		return model.Device{}, errors.New(status.Convert(err).Message())
	}
	var deviceInfo model.Device
	if resp != nil {
		if resp.GetBaseResponse().GetSuccess() {
			deviceInfo.Id = resp.Data.Devices.Id
			deviceInfo.Name = resp.Data.Devices.Name
			deviceInfo.ProductId = resp.Data.Devices.ProductId
			deviceInfo.DeviceSn = resp.Data.Devices.DeviceSn
			deviceInfo.Secret = resp.Data.Devices.Secret
			deviceInfo.External = resp.Data.Devices.External
			deviceInfo.Status = commons.TransformRpcDeviceStatusToModel(resp.Data.Devices.Status)
			deviceInfo.External = resp.Data.Devices.External
			d.deviceCache.Add(deviceInfo)
			return deviceInfo, nil
		} else {
			return deviceInfo, errors.New(resp.GetBaseResponse().GetErrorMessage())
		}
	}
	return deviceInfo, errors.New("unKnow error")
}

func (d *DriverService) getProductProperties(productId string) (map[string]model.Property, bool) {
	return d.productCache.GetProductProperties(productId)
}

func (d *DriverService) getProductPropertyByCode(productId, code string) (model.Property, bool) {
	return d.productCache.GetPropertySpecByCode(productId, code)
}

func (d *DriverService) getProductEvents(productId string) (map[string]model.Event, bool) {
	return d.productCache.GetProductEvents(productId)
}

func (d *DriverService) getProductEventByCode(productId, code string) (model.Event, bool) {
	return d.productCache.GetEventSpecByCode(productId, code)
}

func (d *DriverService) getProductServices(productId string) (map[string]model.Service, bool) {
	return d.productCache.GetProductServices(productId)
}

func (d *DriverService) getProductServiceByCode(productId, code string) (model.Service, bool) {
	return d.productCache.GetServiceSpecByCode(productId, code)
}
