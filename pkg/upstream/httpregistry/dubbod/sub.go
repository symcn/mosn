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
	dubbocommon "github.com/mosn/registry/dubbo/common"
	dubboconsts "github.com/mosn/registry/dubbo/common/constant"
	"net/http"
	"net/url"
)

// subscribe a service from registry
func subscribe(w http.ResponseWriter, r *http.Request) {
	var req subReq
	err := bind(r, &req)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: "subscribe fail, err: " + err.Error()})
		return
	}

	err = doSubUnsub(req, true)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: "subscribe fail, err: " + err.Error()})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "subscribe success"})
}

// unsubscribe a service from registry
func unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req subReq
	err := bind(r, &req)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: "unsubscribe fail, err: " + err.Error()})
		return
	}
	err = doSubUnsub(req, false)

	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: "unsubscribe fail, err: " + err.Error()})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "unsubscribe success"})
}

// map[string]registry.NotifyListener{}
//var dubboInterface2listener = sync.Map{}

func doSubUnsub(req subReq, sub bool) error {
	zookeeperAddr := GetZookeeperAddr()
	var registryPath = registryPathTpl.ExecuteString(map[string]interface{}{
		"addr": zookeeperAddr,
	})
	registryURL, _ := dubbocommon.NewURL(registryPath,
		dubbocommon.WithParams(url.Values{
			dubboconsts.REGISTRY_TIMEOUT_KEY: []string{"5s"},
			dubboconsts.ROLE_KEY:             []string{fmt.Sprint(dubbocommon.CONSUMER)},
		}),
	)

	servicePath := req.Service.Interface // com.mosn.test.UserService
	reg, err := getRegistry(servicePath, dubbocommon.CONSUMER, &registryURL)
	if err != nil {
		return err
	}

	vals := url.Values{
		dubboconsts.ROLE_KEY: []string{fmt.Sprint(dubbocommon.CONSUMER)},
		//dubboconsts.GROUP_KEY: []string{req.Service.Group},
	}
	for k, v := range req.Service.Params {
		vals.Set(k, fmt.Sprint(v))
	}
	dubboURL := dubbocommon.NewURLWithOptions(
		dubbocommon.WithPath(servicePath),
		dubbocommon.WithProtocol(dubbo), // this protocol is used to compare the url, must provide
		dubbocommon.WithParams(vals),
		dubbocommon.WithMethods(req.Service.Methods))

	// register consumer to registry
	if sub {
		err = reg.Register(*dubboURL)
		if err != nil {
			return err
		}
	} else {
		err = reg.UnRegister(*dubboURL)
		if err != nil {
			return err
		}
	}

	//if sub {
	//	// listen to provider change events
	//	var l = &listener{}
	//	go reg.Subscribe(dubboURL, l)
	//	dubboInterface2listener.Store(servicePath, l)
	//
	//	err = addRouteRule(servicePath)
	//	if err != nil {
	//		return err
	//	}
	//} else {
	//	l, ok := dubboInterface2listener.Load(servicePath)
	//	if ok {
	//		err = reg.UnSubscribe(dubboURL, l.(*listener))
	//		if err != nil {
	//			return err
	//		}
	//	}
	//
	//	// NOTICE: router rule will remain in the router manager, but it's ok
	//}

	return nil

}
