package dubbod

import (
	"fmt"
	"net/http"
	"time"

	"mosn.io/mosn/pkg/log"
)

var hb chan struct{}

func init() {
	hb = make(chan struct{}, 3)
	go autoUnpub()
}

func heartbeat(w http.ResponseWriter, r *http.Request) {
	//TODO check some status
	select {
	case hb <- struct{}{}:
	case <-time.After(time.Second * 5):
		response(w, resp{Errno: fail, ErrMsg: "ack fail timeout"})
		return
	}

	fmt.Println(alreadyPublish)
	response(w, resp{Errno: succ, ErrMsg: "ack success", InterfaceList: getInterfaceList()})
}

func autoUnpub() {
	for {
		select {
		case <-time.After(heartBeatExpire * heartBeatNum):
			log.DefaultLogger.Infof("heart beat expire, unpublish all service")
			unpublishAll()
		case <-hb:
			log.DefaultLogger.Infof("heart beat.")
		}
	}
}
