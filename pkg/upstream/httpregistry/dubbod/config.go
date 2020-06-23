package dubbod

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"mosn.io/mosn/pkg/log"
)

const (
	mosnRegistryHttpPortEnvName = "MOSN_REGISTRY_HTTP_PORT"
	defaultHttpPort             = 12181
	zookeeperAddrEnvName        = "MOSN_ZK_ADDRESS"
	defaultZookeeperAddr        = "127.0.0.1:2181"
	zookeeper                   = "zookeeper"
	dubbo                       = "dubbo"
	ip                          = "ip"
	port                        = "port"
	interfaceName               = "interface"
	defaultExportPort           = 20882
	mosnExportDubboPort         = "MOSN_EXPORT_PORT"
	heartBeatExpire             = time.Second * 5
	heartBeatNum                = 3
)

func GetHttpAddr() string {
	httpPort, err := strconv.Atoi(getEnv(mosnRegistryHttpPortEnvName, strconv.Itoa(defaultHttpPort)))
	if err != nil {
		log.DefaultLogger.Fatalf("can not parse http port from env", err.Error())
		return ""
	}
	return fmt.Sprintf(":%d", httpPort)
}

func GetZookeeperAddr() string {
	return getEnv(zookeeperAddrEnvName, defaultZookeeperAddr)
}

func GetExportDubboPort() int {
	port, err := strconv.Atoi(getEnv(mosnExportDubboPort, fmt.Sprintf("%d", defaultExportPort)))
	if err != nil {
		log.DefaultLogger.Fatalf("can not parse export port from env", err.Error())
		return -1
	}
	return port
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
