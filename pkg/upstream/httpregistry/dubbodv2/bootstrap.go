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
	"net/http"
	"time"

	"github.com/go-chi/chi"
	dubbologger "github.com/symcn/registry/dubbo/common/logger"
	"mosn.io/mosn/pkg/admin/store"
	"mosn.io/mosn/pkg/log"
	"mosn.io/pkg/utils"
)

// init the http api for application when application bootstrap
// for sub/pub
func Init() {
	// 1. init router
	r := chi.NewRouter()
	r.Get("/registry/info/sync", registryInfoSyncGet)
	r.Post("/registry/info/sync", registryInfoSync)

	_ = dubbologger.InitLog("./dubbogo.log")

	utils.GoWithRecover(func() {
		for store.GetMosnState() != store.Running {
			log.DefaultLogger.Infof("wait mosn status(%d) running", store.GetMosnState())
			time.Sleep(time.Second * 1)
		}

		if err := http.ListenAndServe(GetHttpAddr(), r); err != nil {
			log.DefaultLogger.Infof("auto write config when updated:%+v", err)
		}
	}, nil)
}
