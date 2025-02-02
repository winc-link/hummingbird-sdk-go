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

package interfaces

import (
	"context"
	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/model"
)

type Driver interface {
	// DeviceNotify 设备增删改通知,删除设备时device参数为空
	DeviceNotify(ctx context.Context, t commons.DeviceNotifyType, deviceId string, device model.Device) error
	// ProductNotify 产品增删改通知,删除产品时product参数为空
	ProductNotify(ctx context.Context, t commons.ProductNotifyType, productId string, product model.Product) error
	// Stop hummingbird-core 服务停止
	Stop(ctx context.Context) error
	// HandlePropertySet 设备属性下发
	HandlePropertySet(ctx context.Context, deviceId string, data model.PropertySet) error
	//HandlePropertyGet 设备属性查询
	HandlePropertyGet(ctx context.Context, deviceId string, data model.PropertyGet) error
	//HandleServiceExecute 设备服务调用
	HandleServiceExecute(ctx context.Context, deviceId string, data model.ServiceExecuteRequest) error
}
