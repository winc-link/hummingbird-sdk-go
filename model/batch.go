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

import "time"

type (
	BatchReport struct {
		CommonRequest `json:",inline"`
		Data          BatchData `json:"data"`
	}
	BatchProperty struct {
		Value interface{} `json:"value"`
		//Time  int64       `json:"time"`
	}
	BatchEvent struct {
		//EventTime    int64                  `json:"eventTime"`
		OutputParams map[string]interface{} `json:"outputParams"`
	}
	BatchData struct {
		Properties map[string]BatchProperty `json:"properties"`
		Events     map[string]BatchEvent    `json:"events"`
	}
)

func NewBatchReport(needACK bool, data BatchData) BatchReport {
	var ack int8
	if needACK {
		ack = 1
	}
	return BatchReport{
		CommonRequest: CommonRequest{
			Version: Version,
			Time:    time.Now().UnixMilli(),
			Sys: ACK{
				Ack: ack,
			},
		},
		Data: data,
	}
}
