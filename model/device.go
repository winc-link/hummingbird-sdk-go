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

package model

import (
	"github.com/winc-link/edge-driver-proto/driverdevice"
	"github.com/winc-link/hummingbird-sdk-go/commons"
)

type (
	Device struct {
		Id           string
		Name         string
		ProductId    string
		DeviceSn     string
		Description  string
		Status       commons.DeviceStatus
		Secret       string
		Ip           string
		Port         string
		Lat          string
		Lon          string
		Location     string
		ParentId     string
		Manufacturer string
		Model        string
		Transport    string
		SlaveId      string
		Period       string
		External     map[string]string
	}
)

type (
	AddDevice struct {
		Name         string               //名称
		ProductId    string               //产品ID
		DeviceSn     string               //设备唯一标识
		Status       commons.DeviceStatus //状态
		Ip           string               // IP
		Port         string               //端口
		Lat          string               //纬度
		Lon          string               //经度
		Location     string               //位置
		ParentId     string               //父设备ID
		Manufacturer string               //工厂
		Model        string               //版本
		Description  string               //描述
		Transport    string               //协议
		External     map[string]string    //附加信息
		//Domain       string
	}

	UpdateDevice struct {
		Id           string
		Name         string
		ProductId    string
		DeviceSn     string
		Status       commons.DeviceStatus
		Ip           string
		Port         string
		Lat          string
		Lon          string
		Location     string
		ParentId     string
		Manufacturer string
		Model        string
		Description  string
		Transport    string
		External     map[string]string
	}
)

func NewAddDevice(name, productId, deviceSn string, status commons.DeviceStatus, ip, port, lat, lon, location, parentId,
	manufacturer, model string, description string, transport string, external map[string]string) AddDevice {
	return AddDevice{
		Name:         name,
		ProductId:    productId,
		DeviceSn:     deviceSn,
		Status:       status,
		Ip:           ip,
		Port:         port,
		Lat:          lat,
		Lon:          lon,
		Location:     location,
		ParentId:     parentId,
		Manufacturer: manufacturer,
		Model:        model,
		Description:  description,
		Transport:    transport,
		External:     external,
	}
}

func NewUpdateDevice(id, name, productId, deviceSn string, status commons.DeviceStatus, ip, port, lat, lon, location,
	parentId, manufacturer, model string, description string, transport string, external map[string]string) UpdateDevice {
	return UpdateDevice{
		Id:           id,
		Name:         name,
		ProductId:    productId,
		DeviceSn:     deviceSn,
		Status:       status,
		Ip:           ip,
		Port:         port,
		Lat:          lat,
		Lon:          lon,
		Location:     location,
		ParentId:     parentId,
		Manufacturer: manufacturer,
		Model:        model,
		Description:  description,
		Transport:    transport,
		External:     external,
	}
}

func TransformDeviceModel(dev *driverdevice.Device) Device {
	var d Device
	d.Id = dev.GetId()
	d.Name = dev.GetName()
	d.ProductId = dev.GetProductId()
	d.DeviceSn = dev.GetDeviceSn()
	d.Description = dev.GetDescription()
	d.Status = commons.TransformRpcDeviceStatusToModel(dev.GetStatus())
	d.Secret = dev.GetSecret()
	d.Ip = dev.GetIp()
	d.Port = dev.GetPort()
	d.Lat = dev.GetLat()
	d.Lon = dev.GetLon()
	d.Location = dev.GetLocation()
	d.ParentId = dev.GetParentId()
	d.Manufacturer = dev.GetManufacturer()
	d.Model = dev.GetModel()
	d.External = dev.GetExternal()
	d.Transport = dev.GetTransport()
	d.SlaveId = dev.GetSlaveId()
	d.Period = dev.GetPeriod()
	return d
}

func UpdateDeviceModelFieldsFromProto(dev *Device, patch *driverdevice.Device) {
	if patch.GetName() != "" {
		dev.Name = patch.GetName()
	}
	if patch.GetProductId() != "" {
		dev.ProductId = patch.GetProductId()
	}
	if patch.GetDeviceSn() != "" {
		dev.DeviceSn = patch.GetDeviceSn()
	}
	if patch.GetDescription() != "" {
		dev.Description = patch.GetDescription()
	}
	if patch.GetStatus().String() != "" {
		dev.Status = commons.TransformRpcDeviceStatusToModel(patch.GetStatus())
	}
	if patch.GetSecret() != "" {
		dev.Secret = patch.GetSecret()
	}
	if patch.GetIp() != "" {
		dev.Ip = patch.GetIp()
	}
	if patch.GetPort() != "" {
		dev.Port = patch.GetPort()
	}
	if patch.GetLat() != "" {
		dev.Lat = patch.GetLat()
	}
	if patch.GetLon() != "" {
		dev.Lon = patch.GetLon()
	}
	if patch.GetLocation() != "" {
		dev.Location = patch.GetLocation()
	}
	if patch.GetParentId() != "" {
		dev.ParentId = patch.GetParentId()
	}
	if patch.GetManufacturer() != "" {
		dev.Manufacturer = patch.GetManufacturer()
	}
	if patch.GetModel() != "" {
		dev.Model = patch.GetModel()
	}
	if patch.GetExternal() != nil {
		dev.External = patch.GetExternal()
	}
	if patch.GetTransport() != "" {
		dev.Transport = patch.GetTransport()
	}
}
