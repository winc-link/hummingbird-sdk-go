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
	"github.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/hummingbird-sdk-go/interfaces"
	"github.com/winc-link/hummingbird-sdk-go/internal/logger"
	"github.com/winc-link/hummingbird-sdk-go/model"
	"gorm.io/gorm"
)

// Start 启动驱动
func (d *DriverService) Start(driver interfaces.Driver) error {
	return d.start(driver)
}

// GetLogger 获取日志接口
func (d *DriverService) GetLogger() logger.Logger {
	return d.logger
}

// GetCustomParam 获取自定义参数
func (d *DriverService) GetCustomParam() string {
	return d.cfg.CustomParam
}

// GetDBClient 获取数据库连接
func (d *DriverService) GetDBClient() *gorm.DB {
	return d.dbClient
}

// Online 设备与平台建立连接
func (d *DriverService) Online(deviceId string) error {
	return d.connectIotPlatform(deviceId)
}

// Offline 设备与平台断开连接
func (d *DriverService) Offline(deviceId string) error {
	return d.disconnectIotPlatform(deviceId)
}

// GetConnectStatus 获取设备连接状态
func (d *DriverService) GetConnectStatus(deviceId string) (commons.DeviceConnectStatus, error) {
	return d.getConnectStatus(deviceId)
}

// CreateDevice 创建设备
func (d *DriverService) CreateDevice(device model.AddDevice) (model.Device, error) {
	return d.createDevice(device)
}

// UpdateDevice 更新设备
func (d *DriverService) UpdateDevice(device model.UpdateDevice) (model.Device, error) {
	return d.updateDevice(device)
}

// DeleteDevice 删除设备
func (d *DriverService) DeleteDevice(deviceId string) error {
	return d.deleteDevice(deviceId)
}

// GetDeviceList 获取所有的设备
func (d *DriverService) GetDeviceList() []model.Device {
	return d.getDeviceList()
}

// GetDeviceById 通过设备id获取设备详情
func (d *DriverService) GetDeviceById(deviceId string) (model.Device, bool) {
	return d.getDeviceById(deviceId)
}

// GetDeviceByDeviceSn 通过设备Sn获取设备详情
func (d *DriverService) GetDeviceByDeviceSn(deviceSn string) (model.Device, bool) {
	return d.getDeviceByDeviceSn(deviceSn)
}

// ProductList 获取当前实例下的所有产品
func (d *DriverService) ProductList() map[string]model.Product {
	return d.productCache.All()
}

// GetProductById 根据产品id获取产品信息
func (d *DriverService) GetProductById(productId string) (model.Product, bool) {
	return d.productCache.SearchById(productId)
}

// GetProductProperties 根据产品id获取产品所有属性信息
func (d *DriverService) GetProductProperties(productId string) (map[string]model.Property, bool) {
	return d.getProductProperties(productId)
}

// GetProductPropertyByCode 根据产品id与code获取属性信息
func (d *DriverService) GetProductPropertyByCode(productId, code string) (model.Property, bool) {
	return d.getProductPropertyByCode(productId, code)
}

// GetProductEvents 根据产品id获取产品所有事件信息
func (d *DriverService) GetProductEvents(productId string) (map[string]model.Event, bool) {
	return d.getProductEvents(productId)
}

// GetProductEventByCode 根据产品id与code获取事件信息
func (d *DriverService) GetProductEventByCode(productId, code string) (model.Event, bool) {
	return d.getProductEventByCode(productId, code)
}

// GetProductServices 根据产品id获取产品所有服务信息
func (d *DriverService) GetProductServices(productId string) (map[string]model.Service, bool) {
	return d.getProductServices(productId)
}

// GetProductServiceByCode 根据产品id与code获取服务信息
func (d *DriverService) GetProductServiceByCode(productId, code string) (model.Service, bool) {
	return d.getProductServiceByCode(productId, code)
}

// PropertyReport 物模型属性上报 如果data参数中的Sys.Ack设置为1，则该方法会同步阻塞等待云端返回结果。
func (d *DriverService) PropertyReport(deviceId string, data model.PropertyReport) (model.CommonResponse, error) {
	return d.propertyReport(deviceId, data)
}

// EventReport 物模型事件上报
func (d *DriverService) EventReport(deviceId string, data model.EventReport) (model.CommonResponse, error) {
	return d.eventReport(deviceId, data)
}

// BatchReport 设备批量上报属性和事件 如果data参数中的Sys.Ack设置为1，则该方法会同步阻塞等待云端返回结果。
// 如非必要，不建议设置Sys.Ack
// 废弃
//func (d *DriverService) BatchReport(deviceId string, data model.BatchReport) (model.CommonResponse, error) {
//	return d.batchReport(deviceId, data)
//}

// PropertyDesiredGet 设备拉取属性期望值 如果data参数中的Sys.Ack设置为1，则该方法会同步阻塞等待云端返回结果。
// 废弃
//func (d *DriverService) PropertyDesiredGet(deviceId string, data model.PropertyDesiredGet) (model.PropertyDesiredGetResponse, error) {
//	return d.propertyDesiredGet(deviceId, data)
//}

// PropertyDesiredDelete 设备删除属性期望值 如果data参数中的Sys.Ack设置为1，则该方法会同步阻塞等待云端返回结果。
// 废弃
//func (d *DriverService) PropertyDesiredDelete(deviceId string, data model.PropertyDesiredDelete) (model.PropertyDesiredDeleteResponse, error) {
//	return d.propertyDesiredDelete(deviceId, data)
//}

// PropertySetResponse 设备属性下发响应
func (d *DriverService) PropertySetResponse(deviceId string, data model.PropertySetResponse) error {
	return d.propertySetResponse(deviceId, data)
}

// PropertyGetResponse 设备属性查询响应
func (d *DriverService) PropertyGetResponse(deviceId string, data model.PropertyGetResponse) error {
	return d.propertyGetResponse(deviceId, data)
}

// ServiceExecuteResponse 设备动作执行响应
func (d *DriverService) ServiceExecuteResponse(deviceId string, data model.ServiceExecuteResponse) error {
	return d.serviceExecuteResponse(deviceId, data)
}
