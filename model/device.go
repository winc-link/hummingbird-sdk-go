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
		PrentId      string
		Manufacturer string
		Model        string
		External     map[string]string
	}
)

type (
	AddDevice struct {
		Name         string
		ProductId    string
		DeviceSn     string
		Status       commons.DeviceStatus
		Ip           string
		Port         string
		Lat          string
		Lon          string
		Location     string
		PrentId      string
		Manufacturer string
		Model        string
		External     map[string]string
		Description  string
	}
)

func NewAddDevice(name, productId, deviceSn string, status commons.DeviceStatus, ip, port, lat, lon, location, prentId,
	manufacturer, model string, external map[string]string, description string) AddDevice {
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
		PrentId:      prentId,
		Manufacturer: manufacturer,
		Model:        model,
		External:     external,
		Description:  description,
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
	d.PrentId = dev.GetParentId()
	d.Manufacturer = dev.GetManufacturer()
	d.Model = dev.GetModel()
	d.External = dev.GetExternal()
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
		dev.PrentId = patch.GetParentId()
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
}
