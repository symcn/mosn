package featuregate

import (
	"os"

	"github.com/mosn/holmes"
	"mosn.io/mosn/pkg/log"
)

var (
	appName     = "unknow_app"
	appIP       = "unknow_ip"
	dumpPathPre = "/web/mosn"

	istioMetaWorkloadName = "ISTIO_META_WORKLOAD_NAME"
	istioInstanceIP       = "INSTANCE_IP"
)

type HolmesFeature struct {
	BaseFeatureSpec
}

func (hf *HolmesFeature) InitFunc() {
	log.DefaultLogger.Infof("holmes init")
	if os.Getenv(istioMetaWorkloadName) != "" {
		appName = os.Getenv(istioMetaWorkloadName)
	}
	if os.Getenv(istioInstanceIP) != "" {
		appIP = os.Getenv(istioInstanceIP)
	}
	h, err := holmes.New(
		holmes.WithCollectInterval("5s"),
		holmes.WithCoolDown("1m"),
		holmes.WithDumpPath(dumpPathPre, appName+"_"+appIP+".log"),
		holmes.WithTextDump(),
		holmes.WithCGroup(true),
		holmes.WithLoggerLevel(holmes.LogLevelInfo),

		holmes.WithMemDump(30, 25, 80),
		holmes.WithCPUDump(30, 25, 80),
		holmes.WithGoroutineDump(3000, 25, 200000),
	)
	if err != nil {
		log.DefaultLogger.Fatalf("holmes init err: %+v", err)
	}
	h.EnableMemDump().EnableCPUDump().EnableGoroutineDump()
	go h.Start()
}
