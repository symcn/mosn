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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/mosn/binding"
	dubboreg "github.com/symcn/registry/dubbo"
	dubbocommon "github.com/symcn/registry/dubbo/common"
	dubboconsts "github.com/symcn/registry/dubbo/common/constant"
	zkreg "github.com/symcn/registry/dubbo/zookeeper"
	"github.com/valyala/fasttemplate"
	v2 "mosn.io/mosn/pkg/config/v2"
	routerAdapter "mosn.io/mosn/pkg/router"
)

var (
	dubboPathTpl          = fasttemplate.New("dubbo://{{ip}}:{{port}}/{{interface}}", "{{", "}}")
	registryPathTpl       = fasttemplate.New("registry://{{addr}}", "{{", "}}")
	dubboRouterConfigName = "dubbo" // keep the same with the router config name in mosn_config.json
	registryCacheKey      = "default"
	registrylock          sync.Mutex
)

const (
	succ = iota
	fail
)

// /com.test.cch.UserService --> zk client
var registryClientCache = make(map[string]dubboreg.Registry, 3)

func getRegistry() (dubboreg.Registry, error) {
	var (
		reg dubboreg.Registry
		ok  bool
		err error
	)

	reg, ok = registryClientCache[registryCacheKey]
	if ok {
		return reg, nil
	}

	registrylock.Lock()
	defer registrylock.Unlock()

	reg, ok = registryClientCache[registryCacheKey]
	if ok {
		return reg, nil
	}

	addrStr := GetZookeeperAddr()
	addresses := strings.Split(addrStr, ",")
	address := addresses[0]
	var registryPath = registryPathTpl.ExecuteString(map[string]interface{}{
		"addr": address,
	})

	registryURL, err := dubbocommon.NewURL(registryPath,
		dubbocommon.WithParams(url.Values{
			dubboconsts.REGISTRY_KEY:         []string{zookeeper},
			dubboconsts.REGISTRY_TIMEOUT_KEY: []string{GetZookeeperTimeout()},
			dubboconsts.ROLE_KEY:             []string{fmt.Sprint(dubbocommon.PROVIDER)},
		}),
		dubbocommon.WithLocation(addrStr),
	)
	// init registry
	reg, err = zkreg.NewZkRegistry(&registryURL)
	// store registry object to global cache
	if err == nil {
		registryClientCache[registryCacheKey] = reg
	}
	return reg, err
}

func response(w http.ResponseWriter, respBody interface{}) {
	bodyBytes, err := json.Marshal(respBody)
	if err != nil {
		_, _ = w.Write([]byte("response marshal failed, err: " + err.Error()))
	}
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	_, _ = w.Write(bodyBytes)
}

// bind the struct content from http.Request body/uri
func bind(r *http.Request, data interface{}) error {
	b := binding.Default(r.Method, r.Header.Get("Content-Type"))
	return b.Bind(r, data)
}

var dubboInterface2registerFlag = sync.Map{}

// add a router rule to router manager, avoid duplicate rules
func addRouteRule(servicePath string) error {
	// if already route rule of this service is already added to router manager
	// then skip
	if _, ok := dubboInterface2registerFlag.Load(servicePath); ok {
		return nil
	}

	dubboInterface2registerFlag.Store(servicePath, struct{}{})
	return routerAdapter.GetRoutersMangerInstance().AddRoute(dubboRouterConfigName, "*", &v2.Router{
		RouterConfig: v2.RouterConfig{
			Match: v2.RouterMatch{
				Headers: []v2.HeaderMatcher{
					{
						Name:  "service", // use the xprotocol header field "service"
						Value: servicePath,
					},
				},
			},
			Route: v2.RouteAction{
				RouterActionConfig: v2.RouterActionConfig{
					ClusterName: servicePath,
				},
			},
		},
	})
}
