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

package client

import (
	"context"
	"errors"
	"github.com/winc-link/edge-driver-proto/thingmodel"
	"github.com/winc-link/hummingbird-sdk-go/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/keepalive"
	"time"

	cloudinstanceproto "github.com/winc-link/edge-driver-proto/cloudinstance"
	"github.com/winc-link/edge-driver-proto/custommqttmessage"
	"github.com/winc-link/edge-driver-proto/drivercommon"
	deviceproto "github.com/winc-link/edge-driver-proto/driverdevice"
	productproto "github.com/winc-link/edge-driver-proto/driverproduct"
	driverstorage "github.com/winc-link/edge-driver-proto/driverstorge"
	gatewayproto "github.com/winc-link/edge-driver-proto/gateway"
)

type ResourceClient struct {
	address string
	Conn    *grpc.ClientConn
	drivercommon.CommonClient
	deviceproto.RpcDeviceClient
	gatewayproto.RpcGatewayClient
	productproto.RpcProductClient
	cloudinstanceproto.CloudInstanceServiceClient
	thingmodel.ThingModelUpServiceClient
	custommqttmessage.RpcCustomMqttMessageClient
	driverstorage.DriverStorageClient
}

var connParams = grpc.ConnectParams{
	Backoff: backoff.Config{
		BaseDelay:  time.Second * 1.0,
		Multiplier: 1.0,
		Jitter:     0,
		MaxDelay:   10 * time.Second,
	},
	MinConnectTimeout: time.Second * 3,
}

var keep = keepalive.ClientParameters{
	Time:                10 * time.Second,
	Timeout:             3 * time.Second,
	PermitWithoutStream: true,
}

func dial(cfg config.ClientInfo) (*grpc.ClientConn, error) {
	var (
		err         error
		conn        *grpc.ClientConn
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
	)
	defer cancel()
	if conn, err = grpc.DialContext(ctx, cfg.Address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithKeepaliveParams(keep), grpc.WithConnectParams(connParams)); err != nil {
		return nil, err
	}
	return conn, nil
}

func NewCoreClient(cfg config.ClientInfo) (*ResourceClient, error) {
	var (
		err  error
		conn *grpc.ClientConn
		rc   *ResourceClient
	)
	if cfg.Address == "" {
		return nil, errors.New("required address")
	}
	if conn, err = dial(cfg); err != nil {
		return nil, err
	}
	rc = &ResourceClient{
		address:                    cfg.Address,
		Conn:                       conn,
		CommonClient:               drivercommon.NewCommonClient(conn),
		RpcDeviceClient:            deviceproto.NewRpcDeviceClient(conn),
		RpcGatewayClient:           gatewayproto.NewRpcGatewayClient(conn),
		RpcProductClient:           productproto.NewRpcProductClient(conn),
		CloudInstanceServiceClient: cloudinstanceproto.NewCloudInstanceServiceClient(conn),
		ThingModelUpServiceClient:  thingmodel.NewThingModelUpServiceClient(conn),
		RpcCustomMqttMessageClient: custommqttmessage.NewRpcCustomMqttMessageClient(conn),
		DriverStorageClient:        driverstorage.NewDriverStorageClient(conn),
	}
	return rc, nil
}

func (c *ResourceClient) Close() error {
	return c.Conn.Close()
}

type ThingModelReportClient struct {
	Conn *grpc.ClientConn
	drivercommon.CommonClient
	thingmodel.ThingModelUpServiceClient
}

func NewThingModelReportClient(cfg config.ClientInfo) (*ThingModelReportClient, error) {
	var (
		err  error
		conn *grpc.ClientConn
	)
	if cfg.Address == "" {
		return nil, errors.New("required address")
	}
	if conn, err = dial(cfg); err != nil {
		return nil, err
	}
	return &ThingModelReportClient{
		Conn:                      conn,
		CommonClient:              drivercommon.NewCommonClient(conn),
		ThingModelUpServiceClient: thingmodel.NewThingModelUpServiceClient(conn),
	}, nil
}

func (c *ThingModelReportClient) Close() error {
	return c.Conn.Close()
}
