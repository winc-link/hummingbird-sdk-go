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

package commons

import "github.com/winc-link/edge-driver-proto/driverproduct"

type ProductNodeType string

const (
	NodeTypeUnKnow    ProductNodeType = "unKnow"
	NodeTypeGateway   ProductNodeType = "gateway"
	NodeTypeDevice    ProductNodeType = "device"
	NodeTypeSubDevice ProductNodeType = "subDevice"
)

func TransformRpcNodeTypeToModel(nodeType driverproduct.ProductNodeType) ProductNodeType {
	switch nodeType {
	case driverproduct.ProductNodeType_UnKnow:
		return NodeTypeUnKnow
	case driverproduct.ProductNodeType_Gateway:
		return NodeTypeGateway
	case driverproduct.ProductNodeType_Device:
		return NodeTypeDevice
	case driverproduct.ProductNodeType_SubDevice:
		return NodeTypeSubDevice
	default:
		return NodeTypeUnKnow
	}
}

type ProductNetType string

const (
	NetTypeOther    ProductNetType = "other"
	NetTypeCellular ProductNetType = "cellular"
	NetTypeWifi     ProductNetType = "wifi"
	NetTypeEthernet ProductNetType = "ethernet"
	NetTypeNB       ProductNetType = "nb"
)

func TransformRpcNetTypeToModel(netType driverproduct.ProductNetType) ProductNetType {
	switch netType {
	case driverproduct.ProductNetType_Other:
		return NetTypeOther
	case driverproduct.ProductNetType_Cellular:
		return NetTypeCellular
	case driverproduct.ProductNetType_Wifi:
		return NetTypeWifi
	case driverproduct.ProductNetType_Ethernet:
		return NetTypeEthernet
	case driverproduct.ProductNetType_NB:
		return NetTypeNB
	default:
		return NetTypeOther
	}
}

type ProductProtocolType string

const (
	ProtocolTypeTcp       ProductProtocolType = "TCP"
	ProtocolTypeUdp       ProductProtocolType = "UDP"
	ProtocolTypeCoap      ProductProtocolType = "Coap"
	ProtocolTypeHttp      ProductProtocolType = "HTTP"
	ProtocolTypeWebSocket ProductProtocolType = "WebSocket"
	ProtocolTypeModbusTCP ProductProtocolType = "ModbusTCP"
	ProtocolTypeGB28181   ProductProtocolType = "GB28181"
	ProtocolTypeSiemensS7 ProductProtocolType = "SiemensS7"
	ProtocolTypeIEC104    ProductProtocolType = "IEC104"
	ProtocolTypeUnknow    ProductProtocolType = "Unknow"
)

func TransformRpcProtocolToModel(protocol driverproduct.ProtocolType) ProductProtocolType {
	switch protocol {
	case driverproduct.ProtocolType_TCP:
		return ProtocolTypeTcp
	case driverproduct.ProtocolType_UDP:
		return ProtocolTypeUdp
	case driverproduct.ProtocolType_CoAP:
		return ProtocolTypeCoap
	case driverproduct.ProtocolType_HTTP:
		return ProtocolTypeHttp
	case driverproduct.ProtocolType_WebSocket:
		return ProtocolTypeWebSocket
	case driverproduct.ProtocolType_ModbusTCP:
		return ProtocolTypeModbusTCP
	case driverproduct.ProtocolType_GB28181:
		return ProtocolTypeGB28181
	case driverproduct.ProtocolType_SiemensS7:
		return ProtocolTypeSiemensS7
	case driverproduct.ProtocolType_IEC104:
		return ProtocolTypeIEC104
	default:
		return ProtocolTypeUnknow
	}
}
