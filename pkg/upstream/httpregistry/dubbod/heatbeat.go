package dubbod

import (
	"net/http"
	"time"

	"github.com/symcn/registry/dubbo/common"
	"mosn.io/mosn/pkg/log"
)

var (
	hb            chan struct{}
	expireTime    time.Duration
	timer         *time.Timer
	sendHBTimeout = time.Millisecond * 50
)

func init() {
	hb = make(chan struct{}, 3)

	expireTime = GetHeartExpireTime()
	timer = time.NewTimer(expireTime)

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
	case <-time.After(sendHBTimeout):
		response(w, resp{Errno: fail, ErrMsg: "ack fail timeout"})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "ack success", PubInterfaceList: getPubInterfaceList(), SubInterfaceList: getSubInterfaceList()})
}

func loopCheckHeartbeat() {
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.DefaultLogger.Infof("heartbeat expire, unPublish unSub all service")

			go unPublishAll()
			go unSubAll()

			timer.Reset(expireTime)
		}
	}
}

func autoUnPub() {
	for {
		select {
		case <-hb:
			timer.Reset(expireTime)
			log.DefaultLogger.Debugf("heartbeat ack succ.")
		}
	}
}
