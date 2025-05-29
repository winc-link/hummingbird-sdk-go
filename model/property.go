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

type (
	PropertyData struct {
		Value interface{} `json:"value"` // 上报的属性值
		Time  int64       `json:"time"`  // 属性变更时间戳
	}

	// PropertyReport 属性上报 属性查询响应
	PropertyReport struct {
		CommonRequest `json:",inline"`
		Data          map[string]interface{} `json:"data"`
	}

	// PropertySet 属性下发
	PropertySet struct {
		CommonRequest `json:",inline"`
		Data          map[string]interface{} `json:"data"`
		Spec          map[string]Property    `json:"-"`
	}

	// PropertyGet 属性查询
	PropertyGet struct {
		CommonRequest `json:",inline"`
		Data          []string            `json:"data"`
		Spec          map[string]Property `json:"-"`
	}
	// PropertyGetResponse 属性查询设备响应
	PropertyGetResponse struct {
		//CommonResponse `json:",inline"`
		MsgId string                    `json:"msgId"`
		Data  []PropertyGetResponseData `json:"data"`
	}
	PropertyGetResponseData struct {
		Code  string      `json:"code"`
		Value interface{} `json:"value"`
		Time  int64       `json:"time"`
	}

	PropertySetResponse struct {
		MsgId string                  `json:"msgId"`
		Data  PropertySetResponseData `json:"data"`
	}
	PropertySetResponseData struct {
		ErrorMessage string `json:"errorMessage"`
		Code         uint32 `json:"code"`
		Success      bool   `json:"success"`
	}

	// PropertyDesiredGet 设备拉取属性期望值
	PropertyDesiredGet struct {
		CommonRequest `json:",inline"`
		Data          []string `json:"data"`
	}

	// PropertyDesiredGetResponse 设备拉取属性期望值响应
	PropertyDesiredGetResponse struct {
		CommonResponse `json:",inline"`
		Data           map[string]PropertyDesiredGetValue `json:"data"`
	}

	// PropertyDesiredDelete 设备清除属性期望值
	PropertyDesiredDelete struct {
		CommonRequest `json:",inline"`
		Data          map[string]PropertyDesiredDeleteValue `json:"data"`
	}

	// PropertyDesiredDeleteResponse 设备清除属性期望值响应
	PropertyDesiredDeleteResponse struct {
		CommonResponse `json:",inline"`
		Data           map[string]PropertyDesiredGetValue `json:"data"`
	}

	PropertyDesiredGetValue struct {
		Value   interface{} `json:"value"`
		Version int64       `json:"version"`
	}

	PropertyDesiredDeleteValue struct {
		Version int64 `json:"version"`
	}

	EventBusData struct {
		T           int64                  `json:"t"`
		MsgId       string                 `json:"msgId"`
		DeviceId    string                 `json:"deviceId"`
		ProductId   string                 `json:"productId"`
		MessageType string                 `json:"messageType"`
		Data        map[string]interface{} `json:"data"`
	}
)

func NewPropertyReport(commonRequest CommonRequest, data map[string]interface{}) PropertyReport {
	return PropertyReport{
		CommonRequest: commonRequest,
		Data:          data,
	}
}

//func NewPropertyReport(ack bool, data map[string]interface{}) PropertyReport {
//	var needAck int8
//	if ack {
//		needAck = 1
//	}
//	return PropertyReport{
//		CommonRequest: CommonRequest{
//			Time:    time.Now().UnixMilli(),
//			Version: Version,
//			Sys: ACK{
//				Ack: needAck,
//			},
//		},
//		Data: data,
//	}
//}

func NewPropertyGetResponse(msgId string, data []PropertyGetResponseData) PropertyGetResponse {
	return PropertyGetResponse{
		MsgId: msgId,
		Data:  data,
	}
}

func NewPropertySetResponse(msgId string, data PropertySetResponseData) PropertySetResponse {
	return PropertySetResponse{
		MsgId: msgId,
		Data:  data,
	}
}
