package dubbodv2

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
	zkOperatorInterval          = 1
	zkOperatorIntervalKey       = "MOSN_ZK_OPERATOR_INTERVAL"
	autoCheckNum                = 10
	autoCheckNumKey             = "AUTO_CHECK_REGISTRY_INFO_NUM"
	autoCheckInterval           = 60
	autoCheckIntervalKey        = "AUTO_CHECK_INTERVAL"

	// if is center, mosn will use request host and port
	// if not use request host and MOSN_EXPORT_PORT
	isCenterKey = "MOSN_CENTER_MODE"

	// Path{dubbo://:@10.12.214.61:20882/?interface=abc\u0026group=\u0026version=} has been registered
	zkNodeHasBeenRegisteredErr = "already registered"
	zkNodeHasNotRegisteredErr  = "has not registered"
	zkNodeNotExistErr          = "node does not exist"
	zkConnErr                  = fmt.Errorf("zk not connected")
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
		return defaultExportPort
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

func GetZkInterval() time.Duration {
	et, err := strconv.Atoi(getEnv(zkOperatorIntervalKey, fmt.Sprintf("%d", zkOperatorInterval)))
	if err != nil || et < 1 {
		return time.Second * time.Duration(zkOperatorInterval)
	}
	return time.Second * time.Duration(et)
}

// GetAutoCheckNum auto check num
// >0 check limit n
// =0 no check
// <0 check with not limit
func GetAutoCheckNum() int {
	acn, err := strconv.Atoi(getEnv(autoCheckNumKey, fmt.Sprintf("%d", autoCheckNum)))
	if err != nil {
		return autoCheckNum
	}
	return acn
}

func GetAutoCheckInterval() time.Duration {
	et, err := strconv.Atoi(getEnv(autoCheckIntervalKey, fmt.Sprintf("%d", autoCheckInterval)))
	if err != nil || et < 1 {
		return time.Second * time.Duration(autoCheckInterval)
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
