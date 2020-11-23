/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dubbodm

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	"mosn.io/mosn/pkg/types"
)

const (
	ProtocolName = "dubbodm"
)

// dubbo protocol
const (
	HeaderLen   = 16
	IdLen       = 8
	MagicIdx    = 0
	FlagIdx     = 2
	StatusIdx   = 3
	IdIdx       = 4
	DataLenIdx  = 12
	DataLenSize = 4
)

// req/resp type
const (
	CmdTypeResponse      byte   = 0 // cmd type
	CmdTypeRequest       byte   = 1
	CmdTypeRequestOneway byte   = 2
	UnKnownCmdType       string = "unknown cmd type"
)

/*
 * https://dubbo.apache.org/zh-cn/blog/dubbo-protocol.html
 * 20 - OK
 * 30 - CLIENT_TIMEOUT
 * 31 - SERVER_TIMEOUT
 * 40 - BAD_REQUEST
 * 50 - BAD_RESPONSE
 * 60 - SERVICE_NOT_FOUND
 * 70 - SERVICE_ERROR
 * 80 - SERVER_ERROR
 * 90 - CLIENT_ERROR
 * 100 - SERVER_THREADPOOL_EXHAUSTED_ERROR
 */
const (
	RespStatusOK                             = 20
	RespStatusClientTimeout                  = 30
	RespStatusServerTimeout                  = 31
	RespStatusBadRequest                     = 40
	RespStatusBadResponse                    = 50
	RespStatusServiceNotFound                = 60
	RespStatusServiceError                   = 70
	RespStatusServerError                    = 80
	RespStatusClientError                    = 90
	RespStatusServerThreadpoolExhaustedError = 100
)

const (
	EventRequest  int = 1 // request
	EventResponse int = 0 // response
)

const (
	FrameworkVersionNameHeader string = "dubbo"
	ServiceNameHeader          string = "service"
	MethodNameHeader           string = "method"
	VersionNameHeader          string = "version"
	GroupNameHeader            string = "group"
	InterfaceNameHeader        string = "interface"
	PathNameHeadere            string = "path"
)

const (
	EgressDubbo  string = "egress_dubbo"
	IngressDubbo string = "ingress_dubbo"
)

const (
	ResponseStatusSuccess uint16 = 0x14 // 0x14 response status
)

type dubboStatusInfo struct {
	Status byte
	Msg    string
}

var (
	dubboMosnStatusMap = map[int]dubboStatusInfo{
		types.CodecExceptionCode:    {Status: hessian.Response_SERVICE_ERROR, Msg: "0|codec exception"},
		types.UnknownCode:           {Status: hessian.Response_SERVICE_ERROR, Msg: "2|unknown"},
		types.DeserialExceptionCode: {Status: hessian.Response_SERVICE_ERROR, Msg: "3|deserial exception"},
		types.SuccessCode:           {Status: hessian.Response_OK, Msg: "200|success"},
		types.PermissionDeniedCode:  {Status: hessian.Response_SERVER_ERROR, Msg: "403|permission denied"},
		types.RouterUnavailableCode: {Status: hessian.Response_SERVICE_NOT_FOUND, Msg: "404|router unavailable"},
		types.InternalErrorCode:     {Status: hessian.Response_SERVICE_ERROR, Msg: "500|internal error"},
		types.NoHealthUpstreamCode:  {Status: hessian.Response_SERVICE_NOT_FOUND, Msg: "502|no health upstream"},
		types.UpstreamOverFlowCode:  {Status: hessian.Response_BAD_REQUEST, Msg: "503|upstream overflow"},
		types.TimeoutExceptionCode:  {Status: hessian.Response_CLIENT_TIMEOUT, Msg: "504|timeout"},
		types.LimitExceededCode:     {Status: hessian.Response_BAD_REQUEST, Msg: "509|limit exceeded"},
	}
)
