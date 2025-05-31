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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
	"github.com/winc-link/hummingbird-sdk-go/datadb/clickhouse"
	"github.com/winc-link/hummingbird-sdk-go/datadb/influxdb"
	"github.com/winc-link/hummingbird-sdk-go/datadb/tdengine"
	"github.com/winc-link/hummingbird-sdk-go/monitor"
	"gorm.io/driver/sqlite"
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
	eventBusClient    eventBus
	baseMessage       commons.BaseMessage
	dbClient          *gorm.DB
	dataDbClient      datadb.DataBase
	readyChan         chan struct{}

	//customConfig
	userDefinedMessageQueueConfig *MessageQueueConnConfig
	userDefinedMetaBasesConfig    *MetaBasesConnConfig
	userDefinedDataBasesConfig    *DataBasesConnConfig
}

type Options func(srv *DriverService)

func WithCustomMessageQueueConfig(config *MessageQueueConnConfig) Options {
	return func(srv *DriverService) {
		srv.userDefinedMessageQueueConfig = config
	}
}

func WithCustomMetaBasesConfig(config *MetaBasesConnConfig) Options {
	return func(srv *DriverService) {
		srv.userDefinedMetaBasesConfig = config
	}
}

func WithCustomDataBasesConfig(config *DataBasesConnConfig) Options {
	return func(srv *DriverService) {
		srv.userDefinedDataBasesConfig = config
	}
}

type MessageQueueConnConfig struct {
	Protocol          string
	Host              string
	Port              uint32
	Type              string
	MessageQueueTopic string
}

func initMessageQueue(hummingbirdConfig *drivercommon.ConfigResponse, config *MessageQueueConnConfig) (mqtt.Client, string, error) {

	var (
		protocol          string
		host              string
		port              uint32
		messageQueueType  string
		messageQueueTopic string

		mqttClient mqtt.Client
	)

	if config == nil {
		protocol = hummingbirdConfig.MessageQueue.Protocol
		host = hummingbirdConfig.MessageQueue.Host
		port = hummingbirdConfig.MessageQueue.Port
		messageQueueType = hummingbirdConfig.MessageQueue.Type
		messageQueueTopic = hummingbirdConfig.MessageQueue.PublishTopicPrefix
	} else {
		protocol = config.Protocol
		host = config.Host
		port = config.Port
		messageQueueType = config.Type
		messageQueueTopic = config.MessageQueueTopic
	}

	switch messageQueueType {
	case constants.MessageQueueMqtt:
		opts := mqtt.NewClientOptions()
		opts.AddBroker(fmt.Sprintf("%s://%s:%d", protocol, host, port))
		opts.SetAutoReconnect(true)
		//opts.SetConnectRetry(true)
		opts.SetConnectRetryInterval(3 * time.Second) // 重连间隔
		opts.SetConnectTimeout(2 * time.Second)
		opts.OnConnect = func(c mqtt.Client) {
			//log.Info("mqtt connect success")
		}
		opts.OnConnectionLost = func(c mqtt.Client, err error) {
			//log.Error("mqtt connect lost")
		}
		mqttClient = mqtt.NewClient(opts)
		if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
			return mqttClient, messageQueueTopic, token.Error()
		}
	}
	return mqttClient, messageQueueTopic, nil

}

// MetaBasesConnConfig 数据库连接
// doc：
// mysql dbs：root:!@#12345678(127.0.0.1:3306)/hummingbird?charset=utf8mb4&parseTime=True&loc=Local
// sqlite dbs：hummingbird/db-data/core-data/core.db?_timeout=5000
type MetaBasesConnConfig struct {
	Type constants.MetadataType
	Dns  string
}

func initMetaBasesDB(hummingbirdConfig *drivercommon.ConfigResponse, basesConfig *MetaBasesConnConfig) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	if basesConfig == nil {
		dsn := hummingbirdConfig.MetaBases.Source
		switch hummingbirdConfig.MetaBases.Type {
		case drivercommon.MetaBasesType_Mysql:
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				return nil, fmt.Errorf("failed to connect to meta bases: %v", err)
			}
		case drivercommon.MetaBasesType_Sqlite:
			db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
			if err != nil {
				return nil, fmt.Errorf("failed to connect to meta bases: %v", err)
			}
		}
		return db, nil
	} else {
		dsn := basesConfig.Dns
		switch basesConfig.Type {
		case constants.MetadataMysql:
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				return nil, fmt.Errorf("failed to connect to meta bases: %v", err)
			}
		case constants.MetadataSqlite:
			db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
			if err != nil {
				return nil, fmt.Errorf("failed to connect to meta bases: %v", err)
			}
		}
		return db, nil
	}
}

type DataBasesConnConfig struct {
	Type       constants.DataBasesType
	Tdengine   tdengine.DbClient
	InfluxDB   influxdb.DbClient
	ClickHouse clickhouse.DbClient
}

func initDataBasesDB(hummingbirdConfig *drivercommon.ConfigResponse, basesConfig *DataBasesConnConfig) (datadb.DataBase, error) {

	var (
		dataBasesType constants.DataBasesType
		td            tdengine.DbClient
		influxDB      influxdb.DbClient
		clickHouse    clickhouse.DbClient
	)

	if basesConfig == nil {
		switch hummingbirdConfig.DataBases.Type {
		case string(constants.DataBasesInfluxdb):
			dataBasesType = constants.DataBasesInfluxdb
			influxDB.Url = hummingbirdConfig.DataBases.InfluxDB.Url
			influxDB.Bucket = hummingbirdConfig.DataBases.InfluxDB.Bucket
			influxDB.Org = hummingbirdConfig.DataBases.InfluxDB.Org
			influxDB.Token = hummingbirdConfig.DataBases.InfluxDB.Token
		case string(constants.DataBasesClickhouse):
			dataBasesType = constants.DataBasesClickhouse
			clickHouse.Addr = hummingbirdConfig.DataBases.ClickHouse.Addr
			clickHouse.Database = hummingbirdConfig.DataBases.ClickHouse.Database
			clickHouse.Username = hummingbirdConfig.DataBases.ClickHouse.Username
			clickHouse.Password = hummingbirdConfig.DataBases.ClickHouse.Password
		case string(constants.DataBasesTdengine):
			dataBasesType = constants.DataBasesTdengine
			td.Dsn = hummingbirdConfig.DataBases.Tdengine.Dsn
		}

	} else {
		dataBasesType = basesConfig.Type
		td = basesConfig.Tdengine
		influxDB = basesConfig.InfluxDB
		clickHouse = basesConfig.ClickHouse
	}

	switch dataBasesType {
	case constants.DataBasesInfluxdb:
		return influxdb.InitClientInfluxDB(influxDB)

	case constants.DataBasesClickhouse:
		return clickhouse.InitClientHouseClient(clickHouse)

	case constants.DataBasesTdengine:
		return tdengine.InitTDengineClient(td)
	default:
		return nil, fmt.Errorf("initDataBasesDB unsupported data type: %s", dataBasesType)
	}
}

func NewDriverService(serviceName string, opts ...Options) *DriverService {

	driverService := &DriverService{}
	var (
		wg              sync.WaitGroup
		err             error
		cfg             *config.DriverConfig
		log             logger.Logger
		coreClient      *client.ResourceClient
		db              *gorm.DB
		mqttClient      mqtt.Client
		mqttClientTopic string

		dataBaseClient datadb.DataBase
	)

	for _, fn := range opts {
		fn(driverService)
	}

	flag.StringVar(&config.FilePath, "c", config.DefaultConfigFilePath, "./driver -c configFile")
	flag.Parse()
	if cfg, err = config.ParseConfig(); err != nil {
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

	hummingbirdConfig, err := coreClient.CommonClient.GetHummingbirdConfig(context.Background(), new(empty.Empty))
	if err != nil {
		log.Errorf("get hummingbird config error: %v", err)
		os.Exit(-1)
	}

	db, err = initMetaBasesDB(hummingbirdConfig, driverService.userDefinedMetaBasesConfig)
	if err != nil {
		log.Errorf("init meta bases db error: %v", err)
		os.Exit(-1)
	}
	mqttClient, mqttClientTopic, err = initMessageQueue(hummingbirdConfig, driverService.userDefinedMessageQueueConfig)
	if err != nil {
		log.Errorf("init message queue error: %v", err)
		os.Exit(-1)
	}
	dataBaseClient, err = initDataBasesDB(hummingbirdConfig, driverService.userDefinedDataBasesConfig)
	if err != nil {
		log.Errorf("init data bases db error: %v", err)
		os.Exit(-1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	driverService = &DriverService{
		ctx:               ctx,
		cancel:            cancel,
		wg:                &wg,
		rpcClient:         coreClient,
		logger:            log,
		cfg:               cfg,
		driverServiceName: serviceName,
		dbClient:          db,
		dataDbClient:      dataBaseClient,
		eventBusClient:    eventBus{client: mqttClient, topic: mqttClientTopic},
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

type eventBus struct {
	client mqtt.Client
	topic  string
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
		d.logger.Errorf("sync device error: %v rpcServer", err)
		os.Exit(-1)
	} else {
		d.deviceCache = deviceCache
	}

	// Sync product
	if productCache, err := cache.InitProductCache(d.baseMessage, d.rpcClient, d.logger); err != nil {
		d.logger.Errorf("sync tsl error: %v rpcServer", err)
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
	monitor.UpQosRequest()
	productId, ok := d.getProductIdByDeviceId(cid)
	if !ok {
		return model.CommonResponse{
			MsgId:        data.MsgId,
			ErrorMessage: constants.ErrorCodeMsgMap[constants.ProductNotFound],
			Code:         constants.ProductNotFound,
			Success:      false,
		}, nil
	}

	err := d.dataDbClient.Insert(context.Background(), constants.DB_PREFIX+cid, data.Data, data.Time)
	if err != nil {
		return model.CommonResponse{
			MsgId:        data.MsgId,
			ErrorMessage: err.Error(),
			Code:         constants.InsertTimeDbErrCode,
			Success:      false,
		}, err
	}

	d.pushMsgToEventBus(eventBusPropertyPayload(cid, productId, data))
	return model.CommonResponse{
		MsgId:        data.MsgId,
		ErrorMessage: constants.ErrorCodeMsgMap[constants.DefaultSuccessCode],
		Code:         constants.DefaultSuccessCode,
		Success:      true,
	}, nil
}

func (d *DriverService) eventReport(cid string, data model.EventReport) (model.CommonResponse, error) {
	monitor.UpQosRequest()
	//msgId := d.node.GetId().String()
	//data.MsgId = msgId
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
	//msgId := d.node.GetId().String()
	//data.MsgId = msgId
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
	//msgId := d.node.GetId().String()
	//data.MsgId = msgId
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
	//msgId := d.node.GetId().String()
	//data.MsgId = msgId
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

func (d *DriverService) getProductIdByDeviceId(deviceId string) (string, bool) {
	device, ok := d.deviceCache.SearchById(deviceId)
	if !ok {
		return "", false
	}
	return device.ProductId, true
}

func (d *DriverService) getDeviceByDeviceSn(deviceSn string) (model.Device, bool) {
	devices := d.deviceCache.All()
	for _, device := range devices {
		if device.DeviceSn == deviceSn {
			return device, true
		}
	}
	return model.Device{}, false
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
	if addDevice.Status == commons.DeviceOnline {
		reqDevice.Status = driverdevice.DeviceStatus_OnLine
	} else if addDevice.Status == commons.DeviceOffline {
		reqDevice.Status = driverdevice.DeviceStatus_OffLine
	} else {
		reqDevice.Status = driverdevice.DeviceStatus_UnKnowStatus
	}
	reqDevice.Ip = addDevice.Ip
	reqDevice.Port = addDevice.Port
	reqDevice.Lat = addDevice.Lat
	reqDevice.Lon = addDevice.Lon
	reqDevice.Location = addDevice.Location
	reqDevice.ParentId = addDevice.ParentId
	reqDevice.Manufacturer = addDevice.Manufacturer
	reqDevice.Model = addDevice.Model
	reqDevice.Description = addDevice.Description
	reqDevice.Transport = addDevice.Transport
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
			deviceInfo.Status = commons.TransformRpcDeviceStatusToModel(resp.Data.Devices.Status)
			deviceInfo.Ip = resp.Data.Devices.Ip
			deviceInfo.Port = resp.Data.Devices.Port
			deviceInfo.Lat = resp.Data.Devices.Lat
			deviceInfo.Lon = resp.Data.Devices.Lon
			deviceInfo.Location = resp.Data.Devices.Location
			deviceInfo.ParentId = resp.Data.Devices.ParentId
			deviceInfo.Manufacturer = resp.Data.Devices.Manufacturer
			deviceInfo.Model = resp.Data.Devices.Model
			deviceInfo.External = resp.Data.Devices.External
			deviceInfo.Description = resp.Data.Devices.Description
			deviceInfo.Transport = resp.Data.Devices.Transport
			d.deviceCache.Add(deviceInfo)
			return deviceInfo, nil
		} else {
			return deviceInfo, errors.New(resp.GetBaseResponse().GetErrorMessage())
		}
	}
	return deviceInfo, errors.New("unKnow error")
}

func (d *DriverService) updateDevice(updateDevice model.UpdateDevice) (device model.Device, err error) {

	var (
		resp *driverdevice.UpdateDeviceRequestResponse
	)

	if updateDevice.ProductId == "" || updateDevice.Name == "" || updateDevice.DeviceSn == "" || updateDevice.Id == "" {
		err = errors.New("param failed")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	reqDevice := new(driverdevice.UpdateDevice)
	reqDevice.Id = updateDevice.Id
	reqDevice.Name = updateDevice.Name
	reqDevice.ProductId = updateDevice.ProductId
	reqDevice.DeviceSn = updateDevice.DeviceSn
	if updateDevice.Status == commons.DeviceOnline {
		reqDevice.Status = driverdevice.DeviceStatus_OnLine
	} else if updateDevice.Status == commons.DeviceOffline {
		reqDevice.Status = driverdevice.DeviceStatus_OffLine
	} else {
		reqDevice.Status = driverdevice.DeviceStatus_UnKnowStatus
	}
	reqDevice.Ip = updateDevice.Ip
	reqDevice.Port = updateDevice.Port
	reqDevice.Lat = updateDevice.Lat
	reqDevice.Lon = updateDevice.Lon
	reqDevice.Location = updateDevice.Location
	reqDevice.ParentId = updateDevice.ParentId
	reqDevice.Manufacturer = updateDevice.Manufacturer
	reqDevice.Model = updateDevice.Model
	reqDevice.Description = updateDevice.Description
	reqDevice.Transport = updateDevice.Transport
	reqDevice.External = updateDevice.External
	req := driverdevice.UpdateDeviceRequest{
		BaseRequest: d.baseMessage.BuildBaseRequest(),
		Device:      reqDevice,
	}
	if resp, err = d.rpcClient.UpdateDevice(ctx, &req); err != nil {
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
			deviceInfo.Status = commons.TransformRpcDeviceStatusToModel(resp.Data.Devices.Status)
			deviceInfo.Ip = resp.Data.Devices.Ip
			deviceInfo.Port = resp.Data.Devices.Port
			deviceInfo.Lat = resp.Data.Devices.Lat
			deviceInfo.Lon = resp.Data.Devices.Lon
			deviceInfo.Location = resp.Data.Devices.Location
			deviceInfo.ParentId = resp.Data.Devices.ParentId
			deviceInfo.Manufacturer = resp.Data.Devices.Manufacturer
			deviceInfo.Model = resp.Data.Devices.Model
			deviceInfo.External = resp.Data.Devices.External
			deviceInfo.Description = resp.Data.Devices.Description
			deviceInfo.Transport = resp.Data.Devices.Transport
			d.deviceCache.Update(deviceInfo)
			return deviceInfo, nil
		} else {
			return deviceInfo, errors.New(resp.GetBaseResponse().GetErrorMessage())
		}
	}

	return deviceInfo, nil
}

func (d *DriverService) deleteDevice(deviceId string) (err error) {
	if deviceId == "" {
		return errors.New("param failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req := new(driverdevice.DeleteDeviceRequest)
	req.DeviceId = deviceId
	req.BaseRequest = d.baseMessage.BuildBaseRequest()
	_, err = d.rpcClient.DeleteDevice(ctx, req)
	if err != nil {
		return errors.New(status.Convert(err).Message())
	}
	d.deviceCache.RemoveById(deviceId)
	return nil
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

func (d *DriverService) pushMsgToEventBus(payload []byte) {
	if token := d.eventBusClient.client.Publish(d.eventBusClient.topic, 1, false, payload); token.Wait() && token.Error() != nil {
		d.logger.Errorf("pushMsgToEventBus error: %s", token.Error())
	}
	return
}

func eventBusPropertyPayload(deviceId, productId string, report model.PropertyReport) []byte {
	var eventData model.EventBusData
	eventData.T = report.Time
	eventData.MsgId = report.MsgId
	eventData.DeviceId = deviceId
	eventData.ProductId = productId
	eventData.MessageType = constants.EventBusTypePropertyReport
	eventData.Data = report.Data
	b, _ := json.Marshal(eventData)
	return b
}
