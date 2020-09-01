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
package dubbodv2

import "sync"

type ServiceRegistrySnap struct {
	ProviderList []ServiceRegistryInfo `json:"provider_list"`
	ConsumerList []ServiceRegistryInfo `json:"consumer_list"`
}

type ServiceRegistryInfo struct {
	Service Service `json:"service"`
	Host    string  `json:"host,omitempty"`
	Port    int     `json:"port,omitempty"`
}

type Service struct {
	Interface string                 `json:"interface" binding:"required"` // eg. com.mosn.service.DemoService
	Methods   []string               `json:"methods"`                      // eg. GetUser,GetProfile,UpdateName
	Params    map[string]interface{} `json:"params"`
}

// response struct for all requests
type ResponseInfo struct {
	Errno       int         `json:"code"`
	ErrMsg      string      `json:"msg"`
	ServiceList ServiceList `json:"service_list"`
}

type ServiceList struct {
	PubInterfaceList        []string `json:"pub_interface_list,omitempty"`
	SubInterfaceList        []string `json:"sub_interface_list,omitempty"`
	DispatchedInterfaceList []string `json:"dispatched_interface_list"`
	Version                 uint64   `json:"version"`
}

type snipInfo struct {
	l          sync.Mutex
	AlreadyPub map[string]struct{}
	AlreadySub map[string]struct{}
	PubList    map[string]ServiceRegistryInfo
	SubList    map[string]ServiceRegistryInfo
	Version    uint64
}

type Role string
type Operat string

const (
	RoleProvider Role = "provider"
	RoleConsmmer Role = "consumer"

	OpRegistry   Operat = "registry"
	OpUnRegistry Operat = "unregistry"
)

type event struct {
	Role        Role
	Operat      Operat
	ServiceInfo ServiceRegistryInfo
	Version     uint64
}
