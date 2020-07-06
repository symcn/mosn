package dubbo

import (
	"sync"

	gometrics "github.com/rcrowley/go-metrics"
	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/metrics"
	"mosn.io/mosn/pkg/types"
)

const (
	ServiceInfo  = "request_total"
	ResponseSucc = "response_succ_total"
	ResponseFail = "response_fail_total"
	RequestTime  = "request_time"
)

var (
	l            sync.Mutex
	statsFactory = make(map[string]*Stats)
)

type Stats struct {
	RequestServiceInfo gometrics.Counter
	ResponseSucc       gometrics.Counter
	ResponseFail       gometrics.Counter
}

func GetStatus(listener, service, method string) *Stats {
	key := service + "-" + method
	if s, ok := statsFactory[key]; ok {
		return s
	}

	l.Lock()
	defer l.Unlock()
	if s, ok := statsFactory[key]; ok {
		return s
	}

	podl := types.GetPodLabels()
	lables := map[string]string{
		"listener": listener,
		"service":  service,
		"method":   method,
		"subset":   podl["sym-group"],
	}

	mts, err := metrics.NewMetrics("mosn", lables)
	if err != nil {
		log.DefaultLogger.Errorf("create metrics fail: %v", err)
		statsFactory[key] = nil
		return nil
	}
	s := &Stats{
		RequestServiceInfo: mts.Counter(ServiceInfo),
		ResponseSucc:       mts.Counter(ResponseSucc),
		ResponseFail:       mts.Counter(ResponseFail),
	}
	statsFactory[key] = s
	return s
}
