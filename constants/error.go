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

package constants

//type ErrorMessage string

const (
	DefaultSuccess string = ""

	SystemError               string = "system error"
	RpcRequestError           string = "rpc request error"
	ManyRequestsError         string = "too many requests"
	FormatError               string = "the format of result is error"
	DeviceNotFoundError       string = "device not found"
	ProductNotFoundError      string = "product not found"
	ReportDataRangeError      string = "data size is not within the defined range"
	PropertyReportTypeError   string = "data report type error"
	PropertyCodeNotFoundError string = "property code not found"
	EventCodeNotFoundError    string = "event code not found"
	ReportDataLengthError     string = "data length is greater than the defined"
)

//type ErrorCode int

const (
	DefaultSuccessCode int = 200

	SystemErrorCode             int = 10001
	RpcRequestErrorCode         int = 10002
	ManyRequestsErrorCode       int = 10003
	FormatErrorCode             int = 10004
	DeviceNotFound              int = 20001
	ProductNotFound             int = 30001
	ReportDataRangeErrorCode    int = 40001
	PropertyReportTypeErrorCode int = 40002
	PropertyCodeNotFound        int = 40003
	EventCodeNotFound           int = 40004
	ReportDataLengthErrorCode   int = 40005
	InsertTimeDbErrCode         int = 50001
)

var ErrorCodeMsgMap = map[int]string{
	DefaultSuccessCode:          DefaultSuccess,
	SystemErrorCode:             SystemError,
	RpcRequestErrorCode:         RpcRequestError,
	ManyRequestsErrorCode:       ManyRequestsError,
	FormatErrorCode:             FormatError,
	DeviceNotFound:              DeviceNotFoundError,
	ProductNotFound:             ProductNotFoundError,
	ReportDataRangeErrorCode:    ReportDataRangeError,
	PropertyReportTypeErrorCode: PropertyReportTypeError,
	PropertyCodeNotFound:        PropertyCodeNotFoundError,
	EventCodeNotFound:           EventCodeNotFoundError,
	ReportDataLengthErrorCode:   ReportDataLengthError,
}
