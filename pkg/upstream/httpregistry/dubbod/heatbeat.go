package dubbod

import "net/http"

func heartbeat(w http.ResponseWriter, r *http.Request) {
	//TODO check some status
	response(w, resp{Errno: succ, ErrMsg: "ack success"})
}
