package dubbod

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/symcn/registry/dubbo/common"
	"mosn.io/mosn/pkg/log"
)

var (
	hb            chan struct{}
	lastHeartBeat int64
	maxLogTime    int64 = 100
)

func init() {
	hb = make(chan struct{}, 3)
	go loopCheckHeartbeat()
	go autoUnPub()
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	reg, err := getRegistry(common.PROVIDER)
	if err != nil {
		response(w, resp{Errno: fail, ErrMsg: err.Error()})
		return
	}
	if reg == nil {
		response(w, resp{Errno: fail, ErrMsg: "zk cache error"})
		return
	}
	if !reg.ConnectState() {
		response(w, resp{Errno: fail, ErrMsg: "zk not connected."})
		return
	}

	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 50):
		response(w, resp{Errno: fail, ErrMsg: "ack fail timeout"})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "ack success", PubInterfaceList: getPubInterfaceList(), SubInterfaceList: getSubInterfaceList()})
}

func loopCheckHeartbeat() {
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			expireSecond := time.Now().UnixNano() - atomic.LoadInt64(&lastHeartBeat)
			if expireSecond < GetHeartExpireTime().Nanoseconds() {
				continue
			}
			// after maxLogTime(100s) don't print log
			if expireSecond/time.Second.Nanoseconds() <= maxLogTime {
				log.DefaultLogger.Infof("heartbeat expire %d s, unPublish unSub all service", (time.Now().UnixNano()-atomic.LoadInt64(&lastHeartBeat))/(time.Second.Nanoseconds()))
			}

			go unPublishAll()
			go unSubAll()
		}
	}
}

func autoUnPub() {
	for {
		select {
		case <-hb:
			log.DefaultLogger.Debugf("heartbeat ack succ.")
			atomic.StoreInt64(&lastHeartBeat, time.Now().UnixNano())
		}
	}
}
