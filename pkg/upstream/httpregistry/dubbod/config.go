package dubbod

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"mosn.io/mosn/pkg/log"
)

const (
	zookeeper     = "zookeeper"
	dubbo         = "dubbo"
	ip            = "ip"
	port          = "port"
	interfaceName = "interface"

	// Path{dubbo://:@10.12.214.61:20882/?interface=abc\u0026group=\u0026version=} has been registered
	zkNodeHasBeenRegisteredErr = "has been registered"
	zkNodeHasNotRegisteredErr  = "has not registered"

	envPreKey                   = "mosn.io/"
	istioMetajsonAnnotationsKey = "ISTIO_METAJSON_ANNOTATIONS"
)

// ISTIO_METAJSON_ANNOTATIONS={"sidecar.istio.io/inject":"true","sidecar.istio.io/interceptionMode":"NONE"}
var (
	envRegistryHttpPort       = envPreKey + "registryHttpPort"
	envRegistryHttpExpireTime = envPreKey + "registryHttpExpireTime"
	envZkAddr                 = envPreKey + "zkAddr"
	envZkConnTimeout          = envPreKey + "zkConnTimeout"
	envDubboExportPort        = envPreKey + "dubboExportPort"
	envCenterMode             = envPreKey + "centerMode" // if is center, mosn will use request host and port. if not use request host and MOSN_EXPORT_PORT

	istioMetajsonAnnotations = map[string]string{}

	defaultValueMap = map[string]string{
		envRegistryHttpPort:       "12181",
		envRegistryHttpExpireTime: "15",
		envZkAddr:                 "127.0.0.1:2181",
		envZkConnTimeout:          "5",
		envDubboExportPort:        "20882",
		envCenterMode:             "false",
	}
)

func init() {
	data, _ := os.LookupEnv(istioMetajsonAnnotationsKey)
	if data == "" {
		log.DefaultLogger.Warnf("http registry server env is empty, use default config")
		return
	}
	if err := json.Unmarshal([]byte(data), &istioMetajsonAnnotations); err != nil {
		log.DefaultLogger.Warnf("http registry server env [%s] unmarshal failed: %v", data, err)
	}
}

func getConfigWithDefault(k string) string {
	v, ok := istioMetajsonAnnotations[k]
	if !ok || len(v) == 0 {
		return defaultValueMap[k]
	}
	return v
}

func GetHttpAddr() string {
	return getConfigWithDefault(envRegistryHttpPort)
}

func GetZookeeperAddr() string {
	return getConfigWithDefault(envZkAddr)
}

func GetZookeeperTimeout() string {
	et, err := strconv.Atoi(getConfigWithDefault(envZkConnTimeout))
	if err != nil || et < 1 {
		return fmt.Sprintf("%ss", defaultValueMap[envZkConnTimeout])
	}
	return fmt.Sprintf("%ds", et)
}

func GetExportDubboPort() int {
	port, err := strconv.Atoi(getConfigWithDefault(envDubboExportPort))
	if err != nil {
		log.DefaultLogger.Fatalf("can not parse export port from env", err.Error())
		return -1
	}
	return port
}

func GetHeartExpireTime() time.Duration {
	et, err := strconv.Atoi(getConfigWithDefault(envRegistryHttpExpireTime))
	if err != nil || et < 1 {
		et, _ = strconv.Atoi(defaultValueMap[envRegistryHttpExpireTime])
		return time.Second * time.Duration(et)
	}
	return time.Second * time.Duration(et)
}

func IsCenter() bool {
	switch getConfigWithDefault(envCenterMode) {
	case "true", "t":
		return true
	default:
		return false
	}
}
