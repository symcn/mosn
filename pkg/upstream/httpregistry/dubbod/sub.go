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
	"mosn.io/mosn/pkg/types"
)

var (
	subl       sync.Mutex
	alreadySub = make(map[string]subReq)
)

func getSubInterfaceList() []string {
	if len(alreadySub) == 0 {
		return nil
	}
	result := make([]string, 0, len(alreadySub))
	for i := range alreadySub {
		result = append(result, i)
	}
	return result
}

// subscribe a service from registry
func subscribe(w http.ResponseWriter, r *http.Request) {
	var req subReq
	err := bind(r, &req)
	if err != nil {
		log.DefaultLogger.Errorf("subscribe error:%+v", err)
		response(w, resp{Errno: fail, ErrMsg: "subscribe fail, err: " + err.Error()})
		return
	}

	subl.Lock()
	defer subl.Unlock()

	_, ok := alreadySub[req.Service.Interface]
	if ok {
		response(w, resp{Errno: succ, ErrMsg: "subscribe success", SubInterfaceList: getSubInterfaceList()})
		return
	}

	for k, v := range types.GetPodLabels() {
		// ! import: should check need rewrite
		if k == "sym-group" {
			k = "flag"
		}

		// avoid recover params
		if _, ok := req.Service.Params[k]; !ok {
			req.Service.Params[k] = v
		}
	}

	err = doSubUnsub(req, true)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: "subscribe fail, err: " + err.Error()})
		return
	}

	alreadySub[req.Service.Interface] = req

	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 50):
	}

	response(w, resp{Errno: succ, ErrMsg: "subscribe success", SubInterfaceList: getSubInterfaceList()})
}

// unsubscribe a service from registry
func unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req subReq
	err := bind(r, &req)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: "unsubscribe fail, err: " + err.Error()})
		return
	}

	subl.Lock()
	defer subl.Unlock()

	storeReq, ok := alreadySub[req.Service.Interface]
	if !ok {
		response(w, resp{Errno: succ, ErrMsg: "unsubscribe success", SubInterfaceList: getSubInterfaceList()})
		return
	}

	err = doSubUnsub(storeReq, false)
	if err != nil {
		log.DefaultLogger.Errorf("unsubscribe error:%+v", err)
		response(w, resp{Errno: fail, ErrMsg: "unsubscribe fail, err: " + err.Error()})
		return
	}

	delete(alreadySub, req.Service.Interface)

	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 50):
	}

	response(w, resp{Errno: succ, ErrMsg: "unsubscribe success", SubInterfaceList: getSubInterfaceList()})
}

func doSubUnsub(req subReq, sub bool) error {
	reg, err := getRegistry()
	if err != nil {
		return err
	}
	if reg == nil {
		return fmt.Errorf("zk cache error")
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
		err = reg.Register(&dubboURL)
		if err != nil {
			return err
		}
	} else {
		err = reg.UnRegister(&dubboURL)
		if err != nil {
			return err
		}
	}

	return nil
}

func unSubAll() (notUnsub []string) {
	if len(alreadySub) == 0 {
		return nil
	}

	notUnsub = make([]string, 0, len(alreadySub))
	subl.Lock()
	defer subl.Unlock()
	for _, req := range alreadySub {
		if e := doSubUnsub(req, false); e != nil {
			log.DefaultLogger.Errorf("can not unsubscribe service {%s} err:%+v", req.Service.Interface, e.Error())
			notUnsub = append(notUnsub, req.Service.Interface)
		} else {
			log.DefaultLogger.Infof("unsubscribe service {%s} succ", req.Service.Interface)
		}
	}
	alreadySub = make(map[string]subReq)
	return notUnsub
}
