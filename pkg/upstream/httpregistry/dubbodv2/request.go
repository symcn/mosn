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

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/symcn/registry/dubbo/common"
	dubbocommon "github.com/symcn/registry/dubbo/common"
	dubboconsts "github.com/symcn/registry/dubbo/common/constant"
	"mosn.io/mosn/pkg/admin/store"
	v2 "mosn.io/mosn/pkg/config/v2"
	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/trace"
)

var (
	l           sync.RWMutex
	snapPubList = make(map[string]ServiceRegistryInfo)
	snapSubList = make(map[string]ServiceRegistryInfo)
	snapVersion uint64

	snapRegistryReadyPubList []string
	snapRegistryReadySubList []string

	alreadyPubList = make(map[string]struct{})
	alreadySubList = make(map[string]struct{})
)

func getRegistryInterfaceList() ServiceList {
	sl := ServiceList{
		PubInterfaceList: snapRegistryReadyPubList,
		SubInterfaceList: snapRegistryReadySubList,
		Version:          snapVersion,
	}
	if len(sl.SubInterfaceList) == 0 {
		return sl
	}

	// return already can route service list
	store.GetMosnConfigWithCb(store.CfgTypeCluster, func(sobj interface{}) {
		if sobj == nil {
			return
		}
		storeClusters, ok := sobj.(map[string]v2.Cluster)
		if !ok {
			return
		}

		// build can dispatchd interface list
		sl.DispatchedInterfaceList = make([]string, 0, len(snapRegistryReadySubList))

		for _, subSvc := range sl.SubInterfaceList {
			for k := range storeClusters {
				if strings.Contains(strings.ToLower(k), strings.ToLower(subSvc)) {
					sl.DispatchedInterfaceList = append(sl.DispatchedInterfaceList, subSvc)
					break
				}
			}
		}
	})

	return sl
}

func registryInfoSyncGet(w http.ResponseWriter, r *http.Request) {
	response(w, ResponseInfo{Errno: succ, ErrMsg: "get service list succ", ServiceList: getRegistryInterfaceList()})
	return
}

// publish a service to registry
func registryInfoSync(w http.ResponseWriter, r *http.Request) {
	_, err := getRegistryWithCheck(common.PROVIDER)
	if err != nil {
		log.DefaultLogger.Errorf("getRegistryWithCheck error:%+v", err)
		response(w, ResponseInfo{Errno: fail, ErrMsg: err.Error()})
		return
	}

	var req ServiceRegistrySnap
	err = bind(r, &req)
	if err != nil {
		log.DefaultLogger.Errorf("bind requestinfo error:%+v", err)
		response(w, ResponseInfo{Errno: fail, ErrMsg: err.Error()})
		return
	}

	registryReq(&req)

	select {
	case hb <- struct{}{}:
	case <-time.After(sendHBTimeout):
	}

	response(w, ResponseInfo{Errno: succ, ErrMsg: "registry service success", ServiceList: getRegistryInterfaceList()})
	return
}

func doPubUnPub(req ServiceRegistryInfo, pub bool) error {
	reg, err := getRegistryWithCheck(common.PROVIDER)
	if err != nil {
		return err
	}

	executeMap := map[string]interface{}{
		interfaceName: req.Service.Interface,
	}
	if IsCenter() {
		executeMap[ip] = req.Host
		executeMap[port] = fmt.Sprintf("%d", req.Port)
	} else {
		executeMap[ip] = trace.GetIp()
		executeMap[port] = fmt.Sprintf("%d", GetDubboExportPort())
	}

	var dubboPath = dubboPathTpl.ExecuteString(executeMap)
	vals := url.Values{
		dubboconsts.ROLE_KEY: []string{fmt.Sprint(dubbocommon.PROVIDER)},
		//dubboconsts.GROUP_KEY:     []string{req.Service.Group},
		dubboconsts.INTERFACE_KEY: []string{req.Service.Interface},
		//dubboconsts.VERSION_KEY:   []string{req.Service.Version},
	}
	for k, v := range req.Service.Params {
		vals.Set(k, fmt.Sprint(v))
	}
	dubboURL, _ := dubbocommon.NewURL(dubboPath,
		dubbocommon.WithParams(vals),
		dubbocommon.WithMethods(req.Service.Methods))

	if pub {
		// publish this service
		return reg.Register(&dubboURL)
	}

	// unpublish this service
	return reg.UnRegister(&dubboURL)
}

func doSubUnsub(req ServiceRegistryInfo, sub bool) error {
	reg, err := getRegistryWithCheck(common.CONSUMER)
	if err != nil {
		return err
	}

	vals := url.Values{
		dubboconsts.ROLE_KEY: []string{fmt.Sprint(dubbocommon.CONSUMER)},
	}
	for k, v := range req.Service.Params {
		vals.Set(k, fmt.Sprint(v))
	}

	var dubboPath = dubboPathTpl.ExecuteString(map[string]interface{}{
		interfaceName: req.Service.Interface,
		ip:            req.Host,
		port:          fmt.Sprintf("%d", req.Port),
	})

	dubboURL, _ := dubbocommon.NewURL(dubboPath,
		dubbocommon.WithPath(req.Service.Interface),
		dubbocommon.WithProtocol(dubbo), // this protocol is used to compare the url, must provide
		dubbocommon.WithParams(vals),
		dubbocommon.WithMethods(req.Service.Methods),
	)

	// register consumer to registry
	if sub {
		return reg.Register(&dubboURL)
	}
	return reg.UnRegister(&dubboURL)
}
