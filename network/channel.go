package network

import (
	"net/http"
	"sync"
)

type Channel struct {
	lck         *sync.Mutex
	idlngReqsts chan *Request
}

func NewChannel() (chnl *Channel) {
	chnl = &Channel{idlngReqsts: make(chan *Request), lck: &sync.Mutex{}}
	return
}

func (chnl *Channel) NextRequest(w http.ResponseWriter, r *http.Request, shuttingDownListener func(), shuttingDownHost func(), canShutdownEnv bool) (reqst *Request) {
	var rqstchn = make(chan *Request, 1)
	defer func() {
		close(rqstchn)

	}()
	go func(rqstc chan *Request) {
		select {
		case rqst := <-chnl.idlngReqsts:
			rqstc <- rqst
		default:
			rqstc <- nil
		}
	}(rqstchn)
	if reqst = <-rqstchn; reqst == nil {
		reqst = NewRequest(chnl, w, r, shuttingDownListener, shuttingDownHost, canShutdownEnv)
	} else {
		reqst.InitRequest(w, r, shuttingDownListener, shuttingDownHost, canShutdownEnv)
	}
	return
}

func (chnl *Channel) EnqueueRequest(reqst *Request) {
	go func(rqst *Request) {
		chnl.idlngReqsts <- rqst
	}(reqst)
	reqst = nil
}
