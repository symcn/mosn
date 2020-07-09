package dubbod

import (
	"net/http"
	"time"

	"mosn.io/mosn/pkg/log"
)

var hb chan struct{}

func init() {
	hb = make(chan struct{}, 3)
	go autoUnPub()
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	//TODO check some status
	select {
	case hb <- struct{}{}:
	case <-time.After(time.Second * 5):
		response(w, resp{Errno: fail, ErrMsg: "ack fail timeout"})
		return
	}

	response(w, resp{Errno: succ, ErrMsg: "ack success", InterfaceList: getInterfaceList()})
}

func autoUnPub() {
	for {
		select {
		case <-time.After(heartBeatExpire * heartBeatNum):
			log.DefaultLogger.Infof("heartbeat expire, unPublish all service")
			go unPublishAll()
		case <-hb:
			log.DefaultLogger.Debugf("heartbeat.")
		}
	}
}
