package dubbod

import (
	"net/http"
	"time"

	"mosn.io/mosn/pkg/log"
)

var (
	hb chan struct{}
)

func init() {
	hb = make(chan struct{}, 10)
	go autoUnPub()
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	// TODO check some status
	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 500):
		response(w, resp{Errno: fail, ErrMsg: "ack fail timeout"})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "ack success", PubInterfaceList: getPubInterfaceList(), SubInterfaceList: getSubInterfaceList()})
	// // response(w, "ok")
}

func autoUnPub() {
	for {
		t := time.NewTicker(GetHeartExpireTime())
		select {
		case <-t.C:
			log.DefaultLogger.Infof("heartbeat expire, unPublish unSub all service")
			go unPublishAll()
			go unSubAll()
		case <-hb:
			log.DefaultLogger.Debugf("heartbeat.")
		}
		t.Stop()
	}
}
