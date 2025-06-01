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
	"github.com/google/uuid"
	"github.com/winc-link/edge-driver-proto/drivercommon"
	"time"
)

const Version = "2.7"

type ACK struct {
	Ack bool `json:"ack"`
}

type CommonResponse struct {
	MsgId        string `json:"msgId"`
	ErrorMessage string `json:"errorMessage"`
	Code         int    `json:"code"`
	Success      bool   `json:"success"`
}

type CommonRequest struct {
	Version string `json:"version"`
	MsgId   string `json:"msgId"`
	Time    int64  `json:"time"`
	Sys     ACK    `json:"sys"`
}

func NewCommonRequest(msgId string, t int64, ack bool) CommonRequest {
	if t == 0 {
		t = time.Now().UnixMilli()
	}
	if msgId == "" {
		msgId = uuid.New().String()
	}
	return CommonRequest{
		Version: Version,
		MsgId:   msgId,
		Time:    t,
		Sys: ACK{
			Ack: ack,
		},
	}
}

func NewDefaultCommonRequest() CommonRequest {
	return CommonRequest{
		Version: Version,
		MsgId:   uuid.New().String(),
		Time:    time.Now().UnixMilli(),
		Sys: ACK{
			Ack: false,
		},
	}
}

func NewCommonResponse(resp *drivercommon.CommonResponse) CommonResponse {
	return CommonResponse{
		ErrorMessage: resp.GetErrorMessage(),
		//Code:         resp.GetCode(),
		Success: resp.GetSuccess(),
	}
}
