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
	// EventReport 设备向云端上报事件
	EventReport struct {
		CommonRequest `json:",inline"`
		Data          EventData `json:"data"`
	}
	EventData struct {
		EventCode    string                 `json:"code"`
		OutputParams map[string]interface{} `json:"outputParams"`
	}
)

func NewEventData(code string, outputParams map[string]interface{}) EventData {
	return EventData{
		EventCode:    code,
		OutputParams: outputParams,
	}
}
