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
	"github.com/google/uuid"
	"github.com/winc-link/hummingbird-sdk-go/constants"
	"github.com/winc-link/hummingbird-sdk-go/datadb"
	"github.com/winc-link/hummingbird-sdk-go/datadb/clickhouse"
	"github.com/winc-link/hummingbird-sdk-go/datadb/influxdb"
	"github.com/winc-link/hummingbird-sdk-go/datadb/redis"
	"github.com/winc-link/hummingbird-sdk-go/datadb/tdengine"
	"github.com/winc-link/hummingbird-sdk-go/internal/executors"
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
	cfg                       *config.DriverConfig
	hummingbirdCfg            *config.HummingbirdConfig
	driverServiceName         string
	logger                    logger.Logger
	deviceCache               cache.DeviceProvider
	productCache              cache.ProductProvider
	driver                    interfaces.Driver
	rpcClient                 *client.ResourceClient
	rpcServer                 *server.RpcService
	baseMessage               commons.BaseMessage
	dbClient                  *gorm.DB
	dataDbClient              datadb.DataBase
	redisClient               *redis.Client
	propertiesTimeDataBatcher *executors.PropertiesTimeDataBatcher
	eventsTimeDataBatcher     *executors.DeviceEventTimeDataBatcher
	logsTimeDataBatcher       *executors.DeviceLogsTimeDataBatcher
	readyChan                 chan struct{}

	userDefinedMetaBasesConfig *MetaBasesConnConfig
	userDefinedDataBasesConfig *DataBasesConnConfig
	userDefinedRedisConfig     *RedisBasesConnConfig
}

type Options func(srv *DriverService)

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

func WithCustomRedisBasesConfig(config *RedisBasesConnConfig) Options {
	return func(srv *DriverService) {
		srv.userDefinedRedisConfig = config
	}
}

type RedisBasesConnConfig struct {
	Address  string
	Password string
	DB       int
}

func initRedisClient(hummingbirdConfig *drivercommon.ConfigResponse, config *RedisBasesConnConfig) (*redis.Client, error) {
	var (
		address  string
		password string
		db       int
	)

	if config == nil {
		address = hummingbirdConfig.RedisBases.Address
		password = hummingbirdConfig.RedisBases.Password
		db = int(hummingbirdConfig.RedisBases.DB)
	} else {
		address = config.Address
		password = config.Password
		db = int(int32(config.DB))
	}
	//链接redis
	return redis.NewClient(address, password, db)
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
		wg             sync.WaitGroup
		err            error
		cfg            *config.DriverConfig
		hummingbirdCfg *config.HummingbirdConfig
		log            logger.Logger
		coreClient     *client.ResourceClient
		db             *gorm.DB
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

	hummingbirdCfg = initHummingbirdCfg(hummingbirdConfig)

	db, err = initMetaBasesDB(hummingbirdConfig, driverService.userDefinedMetaBasesConfig)
	if err != nil {
		log.Errorf("init meta bases db error: %v", err)
		os.Exit(-1)
	}
	dataBaseClient, err = initDataBasesDB(hummingbirdConfig, driverService.userDefinedDataBasesConfig)
	if err != nil {
		log.Errorf("init data bases db error: %v", err)
		os.Exit(-1)
	}

	redisClient, err := initRedisClient(hummingbirdConfig, driverService.userDefinedRedisConfig)
	if err != nil {
		log.Errorf("init redis client error: %v", err)
		os.Exit(-1)
	}

	propertiesTimeDataBatcher := executors.NewTimeDevicePropertiesDataBatcher(dataBaseClient, log, cfg.GetCustomParam())
	eventsTimeDataBatcher := executors.NewTimeDeviceEventsDataBatcher(dataBaseClient, log, cfg.GetCustomParam())
	logsTimeDataBatcher := executors.NewTimeDeviceLogsDataBatcher(dataBaseClient, log, cfg.GetCustomParam())

	ctx, cancel := context.WithCancel(context.Background())
	driverService = &DriverService{
		ctx:                       ctx,
		cancel:                    cancel,
		wg:                        &wg,
		rpcClient:                 coreClient,
		logger:                    log,
		cfg:                       cfg,
		hummingbirdCfg:            hummingbirdCfg,
		driverServiceName:         serviceName,
		dbClient:                  db,
		dataDbClient:              dataBaseClient,
		redisClient:               redisClient,
		propertiesTimeDataBatcher: propertiesTimeDataBatcher,
		eventsTimeDataBatcher:     eventsTimeDataBatcher,
		logsTimeDataBatcher:       logsTimeDataBatcher,
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

	if err = driverService.syncDeviceInfoToDriver(); err != nil {
		log.Error("syncDeviceInfoToDriver error:", err)
		os.Exit(-1)
	}

	return driverService
}

func initHummingbirdCfg(hummingbirdConfig *drivercommon.ConfigResponse) *config.HummingbirdConfig {
	cfg := &config.HummingbirdConfig{}

	if hummingbirdConfig.GetMetaBases() != nil {
		cfg.MetaBases.Source = hummingbirdConfig.GetMetaBases().Source
		cfg.MetaBases.Type = int32(hummingbirdConfig.GetMetaBases().Type)
	}
	if hummingbirdConfig.GetDataBases() != nil {
		cfg.DataBases.Type = hummingbirdConfig.GetDataBases().Type
	}
	cfg.DataBases.InfluxDB = &config.InfluxDB{}
	if hummingbirdConfig.GetDataBases() != nil && hummingbirdConfig.GetDataBases().GetInfluxDB() != nil {
		cfg.DataBases.InfluxDB.Url = hummingbirdConfig.GetDataBases().GetInfluxDB().Url
		cfg.DataBases.InfluxDB.Org = hummingbirdConfig.GetDataBases().GetInfluxDB().Org
		cfg.DataBases.InfluxDB.Bucket = hummingbirdConfig.GetDataBases().GetInfluxDB().Bucket
		cfg.DataBases.InfluxDB.Token = hummingbirdConfig.GetDataBases().GetInfluxDB().Token
	}
	cfg.DataBases.Tdengine = &config.TDengineSource{}
	if hummingbirdConfig.GetDataBases() != nil && hummingbirdConfig.GetDataBases().GetTdengine() != nil {
		cfg.DataBases.Tdengine.Dsn = hummingbirdConfig.GetDataBases().GetTdengine().Dsn
	}
	cfg.DataBases.ClickHouse = &config.ClickHouse{}
	if hummingbirdConfig.GetDataBases() != nil && hummingbirdConfig.GetDataBases().GetClickHouse() != nil {
		cfg.DataBases.ClickHouse.Addr = hummingbirdConfig.GetDataBases().GetClickHouse().Addr
		cfg.DataBases.ClickHouse.Database = hummingbirdConfig.GetDataBases().GetClickHouse().Database
		cfg.DataBases.ClickHouse.Username = hummingbirdConfig.GetDataBases().GetClickHouse().Username
		cfg.DataBases.ClickHouse.Password = hummingbirdConfig.GetDataBases().GetClickHouse().Password
	}
	if hummingbirdConfig.GetRedisBases() != nil {
		cfg.RedisBases.Address = hummingbirdConfig.RedisBases.Address
		cfg.RedisBases.Password = hummingbirdConfig.RedisBases.Password
		cfg.RedisBases.DB = hummingbirdConfig.RedisBases.DB
	}
	return cfg
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

func (d *DriverService) syncDeviceInfoToDriver() error {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for range ticker.C {
			var (
				err  error
				resp *driverdevice.QueryDeviceListResponse
			)
			c, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()

			if resp, err = d.rpcClient.RpcDeviceClient.QueryDeviceList(c, &driverdevice.QueryDeviceListRequest{
				BaseRequest: d.baseMessage.BuildBaseRequest(),
			}); err != nil {
				return
			}
			if !resp.BaseResponse.Success {
				return
			}
			if resp.Data != nil {
				for _, device := range resp.Data.Devices {
					d.deviceCache.Update(model.TransformDeviceModel(device))
				}
			}
		}
	}()
	return nil
}

type DeviceStatus struct {
	LastReportTime time.Time
}

var (
	deviceStatusMap = sync.Map{} // map[string]*DeviceStatus，线程安全
)

// startDeviceStatusChecker 补偿机制：防止因为异常或漏判导致设备状态卡在“离线”状态
//func (d *DriverService) startDeviceStatusChecker() {
//	ticker := time.NewTicker(5 * time.Minute)
//	go func() {
//		for range ticker.C {
//			for deviceID, dev := range d.deviceCache.All() {
//				if dev.Status == commons.DeviceOffline {
//					// 检查最近是否有上报
//					if v, ok := deviceStatusMap.Load(deviceID); ok {
//						status := v.(*DeviceStatus)
//						if time.Since(status.LastReportTime) <= 5*time.Minute {
//							d.Online(deviceID)
//						}
//					}
//				}
//			}
//		}
//	}()
//}

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
	deviceStatusMap.Store(cid, &DeviceStatus{LastReportTime: time.Now()})
	if data.Time == 0 {
		data.Time = time.Now().UnixMilli()
	}
	if data.MsgId == "" {
		data.MsgId = uuid.New().String()
	}
	// 根据设备ID获取设备信息
	device, ok := d.getDeviceById(cid)
	if !ok {
		return model.CommonResponse{
			MsgId:        data.MsgId,
			ErrorMessage: constants.ErrorCodeMsgMap[constants.DeviceNotFound],
			Code:         constants.DeviceNotFound,
			Success:      false,
		}, nil
	}
	// 根据产品ID获取产品信息
	product, ok := d.GetProductById(device.ProductId)
	if !ok {
		return model.CommonResponse{
			MsgId:        data.MsgId,
			ErrorMessage: constants.ErrorCodeMsgMap[constants.ProductNotFound],
			Code:         constants.ProductNotFound,
			Success:      false,
		}, nil
	}
	//如果产品开启了Lua解析脚本，需要对设备上报过来的值进行二次解析。
	if product.LuaScriptEnable {

	}

	// 把设备最新属性数据放入redis中，以便于页面快速查询
	if err := d.redisClient.UpdateDeviceData(cid, data.Time, data.Data); err != nil {
		return model.CommonResponse{
			MsgId:        data.MsgId,
			ErrorMessage: constants.ErrorCodeMsgMap[constants.RedisWriteErrorCode],
			Code:         constants.RedisWriteErrorCode,
			Success:      false,
		}, err
	}

	// 根据设备ID记录每个设备每日上传多少条数据，以便于做统计（设备消息排行榜、设备历史消息统计）
	if err := d.redisClient.IncrDeviceMsgCount(cid, constants.PropertyMsg); err != nil {
		return model.CommonResponse{
			MsgId:        data.MsgId,
			ErrorMessage: constants.ErrorCodeMsgMap[constants.RedisWriteErrorCode],
			Code:         constants.RedisWriteErrorCode,
			Success:      false,
		}, err
	}

	// 把消息推送到redis消息队列中，后端程序消费。
	_ = d.pushMsgToRedisStream(eventBusPropertyPayload(cid, product.Id, data))

	//通过属性的storageMode字段，找到要存入时许数据库的字段
	storeHistoryKeyMap := make(map[string]struct{})
	for _, property := range product.Properties {
		if property.StorageMode == int64(constants.StoreHistory) {
			storeHistoryKeyMap[property.Code] = struct{}{}
		}
	}
	propertiesData := make(map[string]interface{})
	for code, value := range data.Data {
		if _, ok := storeHistoryKeyMap[code]; ok {
			propertiesData[code] = value
		}
	}
	//放入到异步缓冲对列里面，等待时许数据库批量写入。
	if len(propertiesData) > 0 {
		d.propertiesTimeDataBatcher.AddData(model.BatchInsertPropertyData{
			DeviceID: cid,
			T:        data.Time,
			Data:     propertiesData,
		})
	}

	return model.CommonResponse{
		MsgId:        data.MsgId,
		ErrorMessage: constants.ErrorCodeMsgMap[constants.DefaultSuccessCode],
		Code:         constants.DefaultSuccessCode,
		Success:      true,
	}, nil
}

func (d *DriverService) eventReport(cid string, data model.EventReport) (model.CommonResponse, error) {
	monitor.UpQosRequest()
	deviceStatusMap.Store(cid, &DeviceStatus{LastReportTime: time.Now()})
	if data.Time == 0 {
		data.Time = time.Now().UnixMilli()
	}
	if data.MsgId == "" {
		data.MsgId = uuid.New().String()
	}
	productId, ok := d.getProductIdByDeviceId(cid)
	if !ok {
		return model.CommonResponse{
			MsgId:        data.MsgId,
			ErrorMessage: constants.ErrorCodeMsgMap[constants.ProductNotFound],
			Code:         constants.ProductNotFound,
			Success:      false,
		}, nil
	}
	eventData := make(map[string]interface{})
	for k, v := range data.Data {
		b, _ := json.Marshal(v)
		eventData[k] = string(b)
	}

	if err := d.redisClient.IncrDeviceMsgCount(cid, constants.EventMsg); err != nil {
		return model.CommonResponse{}, err
	}

	d.eventsTimeDataBatcher.AddData(model.BatchInsertEventData{
		DeviceID:  cid,
		T:         data.Time,
		EventType: data.EventType,
		Data:      eventData,
	})

	_ = d.pushMsgToRedisStream(eventBusEventPayload(cid, productId, data))
	return model.CommonResponse{
		MsgId:        data.MsgId,
		ErrorMessage: constants.ErrorCodeMsgMap[constants.DefaultSuccessCode],
		Code:         constants.DefaultSuccessCode,
		Success:      true,
	}, nil
}

func (d *DriverService) deviceLogReport(cid string, data model.DeviceLogReport) (model.CommonResponse, error) {
	d.logsTimeDataBatcher.AddData(model.BatchInsertDeviceLogData{
		T:        data.T,
		DeviceID: cid,
		Data:     data.Data,
	})
	return model.CommonResponse{
		MsgId:        data.Data.MessageId,
		ErrorMessage: constants.ErrorCodeMsgMap[constants.DefaultSuccessCode],
		Code:         constants.DefaultSuccessCode,
		Success:      true,
	}, nil
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
	productId, ok := d.getProductIdByDeviceId(deviceId)
	err := d.dbClient.Table("device").Where("id = ?", deviceId).Updates(map[string]interface{}{
		"status":           constants.DeviceOnline,
		"last_online_time": time.Now().UnixMilli(),
	}).Error
	if err != nil {
		return err
	}
	_ = d.pushMsgToRedisStream(eventBusDeviceStatusPayload(deviceId, productId, constants.DeviceOnlineEventBus))
	device, ok := d.deviceCache.SearchById(deviceId)
	if ok {
		device.Status = commons.DeviceOnline
		d.deviceCache.Update(device)
	}
	return nil
}

func (d *DriverService) disconnectIotPlatform(deviceId string) error {
	productId, ok := d.getProductIdByDeviceId(deviceId)
	err := d.dbClient.Table("device").Where("id = ?", deviceId).Updates(map[string]interface{}{
		"status": constants.DeviceOffline,
	}).Error
	if err != nil {
		return err
	}
	_ = d.pushMsgToRedisStream(eventBusDeviceStatusPayload(deviceId, productId, constants.DeviceOfflineEventBus))
	device, ok := d.deviceCache.SearchById(deviceId)
	if ok {
		device.Status = commons.DeviceOffline
		d.deviceCache.Update(device)
	}
	return nil
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

func (d *DriverService) pushMsgToRedisStream(payload []byte) error {
	return d.redisClient.PushMsgToStream(payload)
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

func eventBusEventPayload(deviceId, productId string, report model.EventReport) []byte {
	var eventData model.EventBusData
	eventData.T = report.Time
	eventData.MsgId = report.MsgId
	eventData.DeviceId = deviceId
	eventData.ProductId = productId
	eventData.MessageType = constants.EventBusTypeEventReport
	eventData.Data = make(map[string]interface{})
	for k, v := range report.Data {
		eventData.Data[k] = v
	}
	b, _ := json.Marshal(eventData)
	return b
}

func eventBusDeviceStatusPayload(deviceId, productId string, status string) []byte {
	payload := make(map[string]interface{})
	payload["deviceId"] = deviceId
	payload["productId"] = productId
	payload["messageType"] = constants.EventBusTypeDeviceStatus
	payload["t"] = time.Now().UnixMilli()
	payload["data"] = map[string]interface{}{
		"status": status,
	}
	b, _ := json.Marshal(payload)
	return b
}
