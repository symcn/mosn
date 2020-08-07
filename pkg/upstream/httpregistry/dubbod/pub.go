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

	dubbocommon "github.com/symcn/registry/dubbo/common"
	dubboconsts "github.com/symcn/registry/dubbo/common/constant"
	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/trace"
	"mosn.io/mosn/pkg/types"
)

var (
	publ           sync.Mutex
	alreadyPublish = make(map[string]pubReq)
)

func getPubInterfaceList() []string {
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
		response(w, resp{Errno: succ, ErrMsg: "publish success", PubInterfaceList: getPubInterfaceList()})
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

	response(w, resp{Errno: succ, ErrMsg: "publish success", PubInterfaceList: getPubInterfaceList()})
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
		response(w, resp{Errno: succ, ErrMsg: "unpub success", PubInterfaceList: getPubInterfaceList()})
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

	response(w, resp{Errno: succ, ErrMsg: "unpub success", PubInterfaceList: getPubInterfaceList()})
	return

}

func doPubUnPub(req pubReq, pub bool) error {
	reg, err := getRegistry()
	if err != nil {
		return err
	}
	if reg == nil {
		return fmt.Errorf("zk cache error")
	}

	executeMap := map[string]interface{}{
		interfaceName: req.Service.Interface,
	}
	if IsCenter() {
		executeMap[ip] = req.Host
		executeMap[port] = fmt.Sprintf("%d", req.Port)
	} else {
		executeMap[ip] = trace.GetIp()
		executeMap[port] = fmt.Sprintf("%d", GetExportDubboPort())
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
