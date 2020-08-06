package dubbod

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"mosn.io/mosn/pkg/log"
)

var (
	mosnRegistryHttpPortEnvName = "MOSN_REGISTRY_HTTP_PORT"
	defaultHttpPort             = 12181
	zookeeperAddrEnvName        = "MOSN_ZK_ADDRESS"
	defaultZookeeperAddr        = "127.0.0.1:2181"
	zookeeperConnectTimeoutName = "MOSN_ZK_TIMEOUT"
	zookeeperConnectTimeout     = 5
	zookeeper                   = "zookeeper"
	dubbo                       = "dubbo"
	ip                          = "ip"
	port                        = "port"
	interfaceName               = "interface"
	defaultExportPort           = 20882
	mosnExportDubboPort         = "MOSN_EXPORT_PORT"
	heartBeatExpire             = 15
	mosnHeartBeatExpireKey      = "MOSN_HEART_EXPIRE"

	// if is center, mosn will use request host and port
	// if not use request host and MOSN_EXPORT_PORT
	isCenterKey = "MOSN_CENTER_MODE"
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

func GetZookeeperTimeout() string {
	et, err := strconv.Atoi(getEnv(zookeeperConnectTimeoutName, fmt.Sprintf("%d", zookeeperConnectTimeout)))
	if err != nil || et < 1 {
		et = zookeeperConnectTimeout
	}
	return fmt.Sprintf("%ds", et)
}

func GetExportDubboPort() int {
	port, err := strconv.Atoi(getEnv(mosnExportDubboPort, fmt.Sprintf("%d", defaultExportPort)))
	if err != nil {
		log.DefaultLogger.Fatalf("can not parse export port from env", err.Error())
		return -1
	}
	return port
}

func GetHeartExpireTime() time.Duration {
	et, err := strconv.Atoi(getEnv(mosnHeartBeatExpireKey, fmt.Sprintf("%d", heartBeatExpire)))
	if err != nil || et < 1 {
		return time.Second * time.Duration(heartBeatExpire)
	}
	return time.Second * time.Duration(et)
}

func IsCenter() bool {
	switch getEnv(isCenterKey, "false") {
	case "true", "t":
		return true
	default:
		return false
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getRealEnv(key string) string {
	value, _ := os.LookupEnv(key)
	return value
}
