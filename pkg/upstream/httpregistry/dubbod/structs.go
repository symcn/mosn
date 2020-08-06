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
package dubbod

// for sub && unsub
type subReq struct {
	Service struct {
		Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
		Methods   []string               `json:"methods"`                      // eg. GetUser,GetProfile,UpdateName
		Params    map[string]interface{} `json:"params"`
	} `json:"service"`
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

// for pub && unpub
type pubReq struct {
	Service struct {
		Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
		Methods   []string               `json:"methods"`                      // eg. GetUser,GetProfile,UpdateName
		Params    map[string]interface{} `json:"params"`
	} `json:"service"`
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

// response struct for all requests
type resp struct {
	Errno            int      `json:"code"`
	ErrMsg           string   `json:"msg"`
	PubInterfaceList []string `json:"pub_interface_list,omitempty"`
	SubInterfaceList []string `json:"sub_interface_list,omitempty"`
}
