package dubbodv2

import (
	"sync/atomic"
	"time"

	dubboreg "github.com/symcn/registry/dubbo"
	"mosn.io/mosn/pkg/log"
)

var (
	hb            chan struct{}
	expireTime    time.Duration
	timer         *time.Timer
	sendHBTimeout = time.Millisecond * 50

	autoCheckDone uint64 = 1
)

func init() {
	hb = make(chan struct{}, 3)

	expireTime = GetHeartExpireTime()
	timer = time.NewTimer(expireTime)

	go loopCheckHeartbeat()
	go autoUnPub()
}

func loopCheckHeartbeat() {
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.DefaultLogger.Infof("heartbeat expire, unPublish unSub all service")

			go unRegistryService()

			timer.Reset(expireTime)
		}
	}
}

func unRegistryService() {
	l.RLock()
	if len(snapPubList) == 0 && len(snapSubList) == 0 {
		l.RUnlock()
		return
	}
	l.RUnlock()

	req := &ServiceRegistrySnap{}
	registryReq(req)
}

func autoUnPub() {
	for {
		select {
		case <-hb:
			log.DefaultLogger.Debugf("heartbeat ack succ.")
			timer.Reset(expireTime)
		}
	}
}

func autoCheckSchedule(reg dubboreg.Registry) {
	if atomic.LoadUint64(&autoCheckDone) != 1 {
		return
	}

	atomic.StoreUint64(&autoCheckDone, 0)

	for i := GetAutoCheckNum(); i != 0; {

		time.Sleep(GetAutoCheckInterval())

		l.RLock()
		if len(snapPubList) == 0 && len(snapSubList) == 0 {
			l.RUnlock()
			continue
		}
		l.RUnlock()

		if !reg.ConnectState() {
			// connect close should re-check
			i = GetAutoCheckNum()
			continue
		}

		i--
		syncWithZkHandler()
	}

	atomic.StoreUint64(&autoCheckDone, 1)
}

func syncWithZkHandler() {
	log.DefaultLogger.Infof("auto check registry info with zk.")

	l.Lock()
	pl := make([]ServiceRegistryInfo, 0, len(snapPubList))
	cl := make([]ServiceRegistryInfo, 0, len(snapSubList))
	for _, req := range snapPubList {
		r := req
		pl = append(pl, r)
	}
	for _, req := range snapSubList {
		r := req
		cl = append(cl, r)
	}
	snapPubList = nil
	snapSubList = nil
	l.Unlock()

	registryReq(&ServiceRegistrySnap{
		ProviderList: pl,
		ConsumerList: cl,
	})
}
