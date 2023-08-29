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
	"gitee.com/winc-link/hummingbird-sdk-go/commons"
	"github.com/winc-link/edge-driver-proto/driverdevice"
)

type (
	Device struct {
		//CreateAt    time.Time
		Id          string
		Name        string
		ProductId   string
		Description string
		Status      commons.DeviceStatus
		Platform    commons.IotPlatform
	}
)

type (
	AddDevice struct {
		Name      string
		ProductId string
	}
)

func TransformDeviceModel(dev *driverdevice.Device) Device {
	var d Device
	d.Id = dev.GetId()
	d.Name = dev.GetName()
	d.ProductId = dev.GetProductId()
	d.Description = dev.GetDescription()
	d.Status = commons.TransformRpcDeviceStatusToModel(dev.GetStatus())
	d.Platform = commons.TransformRpcPlatformToModel(dev.GetPlatform())
	return d
}

func UpdateDeviceModelFieldsFromProto(dev *Device, patch *driverdevice.Device) {
	if patch.GetName() != "" {
		dev.Name = patch.GetName()
	}
	if patch.GetProductId() != "" {
		dev.ProductId = patch.GetProductId()
	}
	if patch.GetDescription() != "" {
		dev.Description = patch.GetDescription()
	}

	if patch.GetDescription() != "" {
		dev.Description = patch.GetDescription()
	}

	if patch.GetStatus().String() != "" {
		dev.Status = commons.TransformRpcDeviceStatusToModel(patch.GetStatus())
	}
	if patch.GetPlatform().String() != "" {
		dev.Platform = commons.TransformRpcPlatformToModel(patch.GetPlatform())
	}
}
