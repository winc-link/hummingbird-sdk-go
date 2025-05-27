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

package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/winc-link/hummingbird-sdk-go/monitor"
	"google.golang.org/grpc/reflection"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/interfaces"
	"github.com/winc-link/hummingbird-sdk-go/internal/cache"
	"github.com/winc-link/hummingbird-sdk-go/internal/client"
	"github.com/winc-link/hummingbird-sdk-go/internal/config"
	"github.com/winc-link/hummingbird-sdk-go/internal/logger"
	"github.com/winc-link/hummingbird-sdk-go/model"

	"github.com/winc-link/edge-driver-proto/cloudinstancecallback"
	"github.com/winc-link/edge-driver-proto/devicecallback"
	"github.com/winc-link/edge-driver-proto/drivercommon"
	"github.com/winc-link/edge-driver-proto/productcallback"
	"github.com/winc-link/edge-driver-proto/thingmodel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RpcService struct {
	thingmodel.UnimplementedThingModelDownServiceServer
	productcallback.UnimplementedProductCallBackServiceServer
	devicecallback.UnimplementedDeviceCallBackServiceServer
	cloudinstancecallback.UnimplementedCloudInstanceCallBackServiceServer
	thingmodel.UnimplementedMessageRateServiceServer

	*CommonRPCServer
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup

	lis net.Listener
	s   *grpc.Server

	deviceProvider  cache.DeviceProvider
	productProvider cache.ProductProvider

	driverProvider interfaces.Driver
	//customMqttMessage interfaces.CustomMqttMessage
	//reportPool     *worker.TMWorkerPool

	logger    logger.Logger
	cli       *client.ResourceClient
	isRunning bool
}

//func (server *RpcService) CloudInstanceStatueCallback(ctx context.Context,
//	request *cloudinstancecallback.CloudInstanceStatueCallbackRequest) (*emptypb.Empty, error) {
//	var notifyType commons.CloudPluginNotifyType
//	if request.GetStatus() == cloudinstance.CloudInstanceStatus_Stop {
//		notifyType = commons.CloudPluginStopNotify
//	} else if request.GetStatus() == cloudinstance.CloudInstanceStatus_Start {
//		notifyType = commons.CloudPluginStartNotify
//	}
//	if err := server.driverProvider.CloudPluginNotify(ctx, notifyType, request.GetCloudInstanceName()); err != nil {
//		return new(emptypb.Empty), status.Errorf(codes.Internal, err.Error())
//	}
//	return new(emptypb.Empty), nil
//}

func (server *RpcService) CreateDeviceCallback(ctx context.Context, request *devicecallback.CreateDeviceCallbackRequest) (*emptypb.Empty, error) {
	server.logger.Info("CreateDeviceCallback:", request.String())

	dev := model.TransformDeviceModel(request.GetData())
	server.deviceProvider.Add(dev)
	if err := server.driverProvider.DeviceNotify(ctx, commons.DeviceAddNotify, dev.Id, dev); err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal, err.Error())
	}
	return new(emptypb.Empty), nil
}

func (server *RpcService) UpdateDeviceCallback(ctx context.Context, request *devicecallback.UpdateDeviceCallbackRequest) (*emptypb.Empty, error) {
	server.logger.Info("UpdateDeviceCallback:", request.String())

	deviceId := request.GetData().GetId()
	if len(deviceId) == 0 {
		return new(emptypb.Empty), fmt.Errorf("")
	}
	dev, ok := server.deviceProvider.SearchById(deviceId)
	if !ok {
		return new(emptypb.Empty), status.Errorf(codes.NotFound, "failed to find device %s", deviceId)
	}
	model.UpdateDeviceModelFieldsFromProto(&dev, request.Data)
	server.deviceProvider.Update(dev)

	if err := server.driverProvider.DeviceNotify(ctx, commons.DeviceUpdateNotify, dev.Id, dev); err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal, err.Error())
	}

	return new(emptypb.Empty), nil
}

func (server *RpcService) DeleteDeviceCallback(ctx context.Context, request *devicecallback.DeleteDeviceCallbackRequest) (*emptypb.Empty, error) {
	server.logger.Info("DeleteDeviceCallback:", request.String())
	id := request.GetDeviceId()
	dev, ok := server.deviceProvider.SearchById(id)
	if !ok {
		server.logger.Errorf("failed to find device %s", id)
		return new(emptypb.Empty), status.Errorf(codes.NotFound, "failed to find device %s", id)
	}
	server.deviceProvider.RemoveById(id)
	if err := server.driverProvider.DeviceNotify(ctx, commons.DeviceDeleteNotify, dev.Id, model.Device{}); err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal, err.Error())
	}
	return new(emptypb.Empty), nil
}

func (server *RpcService) CreateProductCallback(ctx context.Context, request *productcallback.CreateProductCallbackRequest) (*emptypb.Empty, error) {
	server.logger.Info("CreateProductCallback:", request.String())
	product := model.TransformProductModel(request.GetData())
	server.productProvider.Add(product)
	if err := server.driverProvider.ProductNotify(ctx, commons.ProductAddNotify, product.Id, product); err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal, err.Error())
	}
	return new(emptypb.Empty), nil
}

func (server *RpcService) UpdateProductCallback(ctx context.Context, request *productcallback.UpdateProductCallbackRequest) (*emptypb.Empty, error) {
	server.logger.Info("UpdateProductCallback:", request.String())
	productId := request.GetData().GetId()
	if len(productId) == 0 {
		return new(emptypb.Empty), fmt.Errorf("")
	}
	_, ok := server.productProvider.SearchById(productId)
	if !ok {
		return new(emptypb.Empty), status.Errorf(codes.NotFound, "failed to find product %s", productId)
	}
	product := model.TransformProductModel(request.GetData())

	server.productProvider.Update(product)

	if err := server.driverProvider.ProductNotify(ctx, commons.ProductUpdateNotify, productId, product); err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal, err.Error())
	}

	return new(emptypb.Empty), nil
}

func (server *RpcService) DeleteProductCallback(ctx context.Context, request *productcallback.DeleteProductCallbackRequest) (*emptypb.Empty, error) {
	server.logger.Info("DeleteProductCallback:", request.String())
	productId := request.GetProductId()
	product, ok := server.productProvider.SearchById(productId)
	if !ok {
		server.logger.Errorf("failed to find product %s", productId)
		return new(emptypb.Empty), status.Errorf(codes.NotFound, "failed to find device %s", productId)
	}
	server.deviceProvider.RemoveById(productId)
	if err := server.driverProvider.ProductNotify(ctx, commons.ProductDeleteNotify, product.Id, model.Product{}); err != nil {
		return new(emptypb.Empty), status.Errorf(codes.Internal, err.Error())
	}
	return new(emptypb.Empty), nil
}

func (server *RpcService) ThingModelMsgIssue(ctx context.Context, request *thingmodel.ThingModelIssueMsg) (*emptypb.Empty, error) {
	monitor.DownQosRequest()
	deviceId := request.GetDeviceId()
	device, ok := server.deviceProvider.SearchById(deviceId)
	if !ok {
		server.logger.Errorf("can't find cid: %s in local cache", deviceId)
		return new(emptypb.Empty), status.Errorf(codes.NotFound, "can't find cid: %s in local cache", deviceId)
	}
	switch request.GetOperationType() {
	case thingmodel.OperationType_PROPERTY_SET:
		var req model.PropertySet
		if err := decoder(request.GetData(), &req); err != nil {
			server.logger.Errorf("decode data error: %s", err)
			return new(emptypb.Empty), status.Errorf(codes.Internal, "decode data error: %s", err)
		}
		req.Spec = make(map[string]model.Property, len(req.Data))
		for k := range req.Data {
			if ps, ok := server.productProvider.GetPropertySpecByCode(device.ProductId, k); !ok {
				server.logger.Warnf("can't find property(%s) spec in product(%s)", k, device.ProductId)
				continue
			} else {
				req.Spec[k] = ps
			}
		}
		err := server.driverProvider.HandlePropertySet(ctx, deviceId, req)
		if err != nil {
			server.logger.Errorf("handlePropertySet error: %s", err)
			return new(emptypb.Empty), status.Errorf(codes.Unknown, err.Error())
		}
	case thingmodel.OperationType_PROPERTY_GET:
		var req model.PropertyGet
		if err := decoder(request.GetData(), &req); err != nil {
			server.logger.Errorf("decode data error: %s", err)
			return new(emptypb.Empty), status.Errorf(codes.Internal, "decode data error: %s", err)
		}

		req.Spec = make(map[string]model.Property, len(req.Data))
		for _, k := range req.Data {
			if ps, ok := server.productProvider.GetPropertySpecByCode(device.ProductId, k); !ok {
				server.logger.Warnf("can't find property(%s) spec in product(%s)", k, device.ProductId)
				continue
			} else {
				req.Spec[k] = ps
			}
		}
		err := server.driverProvider.HandlePropertyGet(ctx, deviceId, req)
		if err != nil {
			server.logger.Errorf("handlePropertyGet error: %s", err)
			return new(emptypb.Empty), status.Errorf(codes.Unknown, err.Error())
		}
	case thingmodel.OperationType_SERVICE_EXECUTE:
		var req model.ServiceExecuteRequest
		if err := decoder(request.GetData(), &req); err != nil {
			server.logger.Errorf("decode data error: %s", err)
			return new(emptypb.Empty), status.Errorf(codes.Internal, "decode data error: %s", err)
		}

		if action, ok := server.productProvider.GetServiceSpecByCode(device.ProductId, req.Data.Code); !ok {
			server.logger.Warnf("can't find action(%s) spec in product(%s)", req.Data.Code, device.ProductId)
		} else {
			req.Spec = action
		}

		err := server.driverProvider.HandleServiceExecute(ctx, deviceId, req)
		if err != nil {
			server.logger.Errorf("handleActionExecute error: %s", err)
		}
	case thingmodel.OperationType_CUSTOM_MQTT_PUBLISH:
		//server.customMqttMessage.CustomMqttMessage("", request.Data)
	default:
		return new(emptypb.Empty), status.Errorf(codes.InvalidArgument, "unsupported operation type")
	}
	return new(emptypb.Empty), nil
}

func (server *RpcService) MessageRate(ctx context.Context, request *emptypb.Empty) (*thingmodel.MessageRateResponse, error) {
	var resp thingmodel.MessageRateResponse
	for _, record := range monitor.GetQosHistory() {
		resp.MessageRate = append(resp.MessageRate, &thingmodel.MessageRate{
			T:       record.T,
			UpQos:   record.UpQPS,
			DownQos: record.DownQPS,
		})
	}
	return &resp, nil
}

func NewRpcService(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc, cfg config.RPCConfig,
	dc cache.DeviceProvider, pc cache.ProductProvider, driverProvider interfaces.Driver, cli *client.ResourceClient,
	logger logger.Logger) (*RpcService, error) {

	if cfg.Address == "" {
		logger.Error("required rpc address")
		return nil, errors.New("required rpc address")
	}

	lis, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		logger.Errorf("failed to listen: %v", err)
		return nil, err
	}
	var s *grpc.Server
	s = grpc.NewServer(grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}), grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     30 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               3 * time.Second,
	}), grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if e := recover(); e != nil {
				logger.Errorf("%s", debug.Stack())
				err = errors.New(fmt.Sprintf("panic:%v", e))
			}
		}()
		reply, err := handler(ctx, req)
		return reply, err
	}))
	reflection.Register(s)
	return &RpcService{
		CommonRPCServer: NewCommonRPCServer(),
		ctx:             ctx,
		cancel:          cancel,
		wg:              wg,
		lis:             lis,
		s:               s,
		deviceProvider:  dc,
		productProvider: pc,
		driverProvider:  driverProvider,
		cli:             cli,
		logger:          logger,
	}, nil
}

func (server *RpcService) Start() error {
	if server.isRunning {
		server.logger.Warn("the grpc server is running")
		return errors.New("the grpc server is running")
	}
	// register method
	drivercommon.RegisterCommonServer(server.s, server)
	productcallback.RegisterProductCallBackServiceServer(server.s, server)
	devicecallback.RegisterDeviceCallBackServiceServer(server.s, server)
	cloudinstancecallback.RegisterCloudInstanceCallBackServiceServer(server.s, server)
	thingmodel.RegisterThingModelDownServiceServer(server.s, server)
	thingmodel.RegisterMessageRateServiceServer(server.s, server)
	monitor.StartQPSCollector()
	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		<-server.ctx.Done()
		server.logger.Info("Server shutting down")
		_ = server.cli.Conn.Close()
		server.s.Stop()
		server.logger.Info("Server shut down")
	}()

	server.logger.Infof("Server starting ( %s )", server.lis.Addr().String())
	server.logger.Info("Server start success.")

	defer func() {
		server.isRunning = false
	}()
	server.isRunning = true
	err := server.s.Serve(server.lis)
	if err != nil {
		server.logger.Errorf("Server failed: %v", err)
		server.cancel()
	} else {
		server.logger.Info("Server stopped")
	}
	server.wg.Wait()
	return err
}
