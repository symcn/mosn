package dubbod

import (
	"net/http"
	"sync/atomic"
	"time"

	"mosn.io/mosn/pkg/log"
)

var (
	hb            chan struct{}
	lastHeartBeat int64
)

func init() {
	hb = make(chan struct{}, 10)
	go loopCheckHeartbeat()
	go autoUnPub()
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	// TODO check some status
	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 50):
		response(w, resp{Errno: fail, ErrMsg: "ack fail timeout"})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "ack success", PubInterfaceList: getPubInterfaceList(), SubInterfaceList: getSubInterfaceList()})
}

func loopCheckHeartbeat() {
	t := time.NewTicker(GetHeartExpireTime())
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if time.Now().UnixNano()-atomic.LoadInt64(&lastHeartBeat) >= GetHeartExpireTime().Nanoseconds() {
				log.DefaultLogger.Infof("heartbeat expire %d s, unPublish unSub all service", (time.Now().UnixNano()-atomic.LoadInt64(&lastHeartBeat))/(time.Second.Nanoseconds()))
				go unPublishAll()
				go unSubAll()
			}
		}
	}
}

func autoUnPub() {
	for {
		select {
		case <-hb:
			log.DefaultLogger.Debugf("heartbeat.")
			atomic.StoreInt64(&lastHeartBeat, time.Now().UnixNano())
		}
	}
}
