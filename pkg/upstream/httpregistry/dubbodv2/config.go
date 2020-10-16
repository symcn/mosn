package dubbodv2

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"mosn.io/mosn/pkg/log"
)

var (
	zookeeper     = "zookeeper"
	dubbo         = "dubbo"
	ip            = "ip"
	port          = "port"
	interfaceName = "interface"

	// Path{dubbo://:@10.12.214.61:20882/?interface=abc\u0026group=\u0026version=} has been registered
	zkNodeHasBeenRegisteredErr = "already registered" // modify this
	zkNodeHasNotRegisteredErr  = "has not registered"
	zkNodeNotExistErr          = "node does not exist"
	zkConnErr                  = fmt.Errorf("zk not connected")

	// ISTIO_METAJSON_ANNOTATIONS={"sidecar.istio.io/inject":"true","sidecar.istio.io/interceptionMode":"NONE"}
	istioMetajsonAnnotationsKey = "ISTIO_METAJSON_ANNOTATIONS"
	envPreKey                   = "mosn.io/"

	envRegistryHttpPort          = envPreKey + "registryHttpPort"
	envRegistryHttpExpireTime    = envPreKey + "registryHttpExpireTime"
	envRegistryAutoCheckNum      = envPreKey + "registryAutoCheckNum"
	envRegistryAutoCheckInterval = envPreKey + "registryAutoCheckInterval"
	envZkAddr                    = envPreKey + "zkAddr"
	envZkConnTimeout             = envPreKey + "zkConnTimeout"
	envZkOperatorErrorInterval   = envPreKey + "zkOperatorErrorInterval"
	envDubboExportPort           = envPreKey + "dubboExportPort"
	envCenterMode                = envPreKey + "centerMode" // if is center, mosn will use request host and port. if not use request host and MOSN_EXPORT_PORT

	EnvRegistryHttpPortValue          = "0.0.0.0:12181"
	EnvRegistryHttpExpireTimeValue    = "15"
	EnvRegistryAutoCheckNumValue      = "10"
	EnvRegistryAutoCheckIntervalValue = "60"
	EnvZkAddrValue                    = "127.0.0.1:2181"
	EnvZkConnTimeoutValue             = "5"
	EnvZkOperatorErrorIntervalValue   = "1"
	EnvDubboExportPortValue           = "20882"
	EnvCenterMode                     = "false"

	defaultValueMap = map[string]string{
		envRegistryHttpPort:          EnvRegistryHttpPortValue,
		envRegistryHttpExpireTime:    EnvRegistryHttpExpireTimeValue,
		envRegistryAutoCheckNum:      EnvRegistryAutoCheckNumValue,
		envRegistryAutoCheckInterval: EnvRegistryAutoCheckIntervalValue,
		envZkAddr:                    EnvZkAddrValue,
		envZkConnTimeout:             EnvZkConnTimeoutValue,
		envZkOperatorErrorInterval:   EnvZkOperatorErrorIntervalValue,
		envDubboExportPort:           EnvDubboExportPortValue,
		envCenterMode:                EnvCenterMode,
	}

	istioMetajsonAnnotations = map[string]string{}
)

func init() {
	data, _ := os.LookupEnv(istioMetajsonAnnotationsKey)
	if data == "" {
		log.DefaultLogger.Warnf("http registry server env is empty, use default config")
		return
	}
	if err := json.Unmarshal([]byte(data), &istioMetajsonAnnotations); err != nil {
		log.DefaultLogger.Warnf("http registry server env [%s] unmarshal failed: %v", data, err)
		return
	}
	for k, v := range istioMetajsonAnnotations {
		if _, ok := defaultValueMap[k]; ok {
			log.DefaultLogger.Infof("[Annotation] recover stetting %s = %s", k, v)
		}
	}
}

func getConfigWithKey(k string) string {
	v, ok := istioMetajsonAnnotations[k]
	if !ok || len(v) == 0 {
		return defaultValueMap[k]
	}
	return v
}

func GetRegistryHttpPort() string {
	return getConfigWithKey(envRegistryHttpPort)
}

func GetRegistryHttpExpireTime() time.Duration {
	et, err := strconv.Atoi(getConfigWithKey(envRegistryHttpExpireTime))
	if err != nil || et < 1 {
		et, _ = strconv.Atoi(defaultValueMap[envRegistryHttpExpireTime])
		return time.Second * time.Duration(et)
	}
	return time.Second * time.Duration(et)
}

// GetRegistryAutoCheckNum auto check num
// >0 check limit n
// =0 no check
// <0 check with not limit
func GetRegistryAutoCheckNum() int {
	acn, err := strconv.Atoi(getConfigWithKey(envRegistryAutoCheckNum))
	if err != nil {
		acn, _ = strconv.Atoi(defaultValueMap[envRegistryAutoCheckNum])
		return acn
	}
	return acn
}

func GetRegistryAutoCheckIntervalTime() time.Duration {
	et, err := strconv.Atoi(getConfigWithKey(envRegistryAutoCheckInterval))
	if err != nil || et < 1 {
		et, _ = strconv.Atoi(defaultValueMap[envRegistryAutoCheckInterval])
		return time.Second * time.Duration(et)
	}
	return time.Second * time.Duration(et)
}

func GetZkAddr() string {
	return getConfigWithKey(envZkAddr)
}

func GetZkConnTimeoutStr() string {
	et, err := strconv.Atoi(getConfigWithKey(envZkConnTimeout))
	if err != nil || et < 1 {
		return fmt.Sprintf("%ss", defaultValueMap[envZkConnTimeout])
	}
	return fmt.Sprintf("%ds", et)
}

func GetZkOperatorErrorIntervalTime() time.Duration {
	et, err := strconv.Atoi(getConfigWithKey(envZkOperatorErrorInterval))
	if err != nil || et < 1 {
		et, _ = strconv.Atoi(defaultValueMap[envZkOperatorErrorInterval])
		return time.Second * time.Duration(et)
	}
	return time.Second * time.Duration(et)
}

func GetDubboExportPort() int {
	port, err := strconv.Atoi(getConfigWithKey(envDubboExportPort))
	if err != nil {
		log.DefaultLogger.Fatalf("can not parse export port from env", err.Error())
		return -1
	}
	return port
}

func IsCenter() bool {
	switch getConfigWithKey(envCenterMode) {
	case "true", "t":
		return true
	default:
		return false
	}
}
