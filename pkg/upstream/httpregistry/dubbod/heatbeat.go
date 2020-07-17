package dubbod

import (
	"net/http"
	"time"

	"mosn.io/mosn/pkg/log"
)

var (
	hb     chan struct{}
	isMock bool
)

func init() {
	hb = make(chan struct{}, 3)
	go autoUnPub()
}

// !import release not pub this route
func heartbeatMock(w http.ResponseWriter, r *http.Request) {
	isMock = !isMock
	response(w, resp{Errno: succ, ErrMsg: "set mock success", InterfaceList: getInterfaceList()})
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	if isMock {
		response(w, resp{Errno: fail, ErrMsg: "mock exception"})
		return
	}
	//TODO check some status
	select {
	case hb <- struct{}{}:
	case <-time.After(time.Millisecond * 50):
		response(w, resp{Errno: fail, ErrMsg: "ack fail timeout"})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "ack success", InterfaceList: getInterfaceList()})
}

func autoUnPub() {
	for {
		select {
		case <-time.After(GetHeartExpireTime()):
			log.DefaultLogger.Infof("heartbeat expire, unPublish unSub all service")
			go unPublishAll()
			go unSubAll()
		case <-hb:
			log.DefaultLogger.Debugf("heartbeat.")
		}
	}
}
