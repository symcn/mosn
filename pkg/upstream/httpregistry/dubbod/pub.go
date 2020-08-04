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

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	dubbocommon "github.com/mosn/registry/dubbo/common"
	dubboconsts "github.com/mosn/registry/dubbo/common/constant"
	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/trace"
	"mosn.io/mosn/pkg/types"
)

var (
	publ           sync.Mutex
	alreadyPublish = make(map[string]pubReq)
)

func getInterfaceList() []string {
	if len(alreadyPublish) == 0 {
		return nil
	}
	result := make([]string, 0, len(alreadyPublish))
	for i := range alreadyPublish {
		result = append(result, i)
	}
	return result
}

// publish a service to registry
func publish(w http.ResponseWriter, r *http.Request) {
	var req pubReq

	err := bind(r, &req)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: err.Error()})
		return
	}

	publ.Lock()
	defer publ.Unlock()

	_, ok := alreadyPublish[req.Service.Interface]
	if ok {
		response(w, resp{Errno: succ, ErrMsg: "publish success", InterfaceList: getInterfaceList()})
		return
	}

	for k, v := range types.GetPodLabels() {
		if k == "sym-group" {
			k = "flag"
			// req.Service.Params["flag"] = v
			// continue
		}

		// avoid recover params
		if _, ok := req.Service.Params[k]; !ok {
			req.Service.Params[k] = v
		}
	}

	err = doPubUnPub(req, true)
	if err != nil {
		log.DefaultLogger.Errorf("publish error:%+v", err)
		response(w, resp{Errno: fail, ErrMsg: "publish fail, err: " + err.Error()})
		return
	}
	alreadyPublish[req.Service.Interface] = req

	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 50):
	}

	response(w, resp{Errno: succ, ErrMsg: "publish success", InterfaceList: getInterfaceList()})
	return
}

// unpublish user service from registry
func unpublish(w http.ResponseWriter, r *http.Request) {
	var req pubReq
	err := bind(r, &req)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: err.Error()})
		return
	}

	publ.Lock()
	defer publ.Unlock()

	storeReq, ok := alreadyPublish[req.Service.Interface]
	if !ok {
		response(w, resp{Errno: succ, ErrMsg: "unpub success", InterfaceList: getInterfaceList()})
		return
	}

	err = doPubUnPub(storeReq, false)
	if err != nil {
		log.DefaultLogger.Errorf("unpublish error:%+v", err)
		response(w, resp{Errno: fail, ErrMsg: "unpub fail, err: " + err.Error()})
		return
	}
	delete(alreadyPublish, storeReq.Service.Interface)

	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 50):
	}

	response(w, resp{Errno: succ, ErrMsg: "unpub success", InterfaceList: getInterfaceList()})
	return

}

func doPubUnPub(req pubReq, pub bool) error {
	addr := GetZookeeperAddr()
	var registryPath = registryPathTpl.ExecuteString(map[string]interface{}{
		"addr": addr,
	})

	registryURL, err := dubbocommon.NewURL(registryPath,
		dubbocommon.WithParams(url.Values{
			dubboconsts.REGISTRY_KEY:         []string{zookeeper},
			dubboconsts.REGISTRY_TIMEOUT_KEY: []string{"5s"},
			dubboconsts.ROLE_KEY:             []string{fmt.Sprint(dubbocommon.PROVIDER)},
		}),
		dubbocommon.WithLocation(addr),
	)
	if err != nil {
		return err
	}

	// find registry from cache
	registryCacheKey := req.Service.Interface
	reg, err := getRegistry(registryCacheKey, dubbocommon.PROVIDER, &registryURL)
	if err != nil {
		return err
	}

	var dubboPath = dubboPathTpl.ExecuteString(map[string]interface{}{
		ip:            trace.GetIp(),
		port:          fmt.Sprintf("%d", GetExportDubboPort()),
		interfaceName: req.Service.Interface,
	})
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
		return reg.Register(dubboURL)
	}

	// unpublish this service
	return reg.UnRegister(dubboURL)

}

func unPublishAll() (notUnpub []string) {
	if len(alreadyPublish) == 0 {
		return nil
	}

	notUnpub = make([]string, 0, len(alreadyPublish))
	publ.Lock()
	defer publ.Unlock()
	for _, req := range alreadyPublish {
		if e := doPubUnPub(req, false); e != nil {
			log.DefaultLogger.Errorf("can not unpublish service {%s} err:%+v", req.Service.Interface, e.Error())
			notUnpub = append(notUnpub, req.Service.Interface)
		} else {
			log.DefaultLogger.Infof("unpublish service {%s} succ", req.Service.Interface)
		}
	}
	alreadyPublish = make(map[string]pubReq)
	return notUnpub
}
