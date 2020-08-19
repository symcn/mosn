package dubbodv2

import (
	"time"

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
