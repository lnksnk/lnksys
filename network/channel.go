package network

import (
	"net/http"
	"sync"
)

type Channel struct {
	lck         *sync.Mutex
	idlngReqsts []*Request
}

func NewChannel() (chnl *Channel) {
	chnl = &Channel{idlngReqsts: []*Request{}, lck: &sync.Mutex{}}
	return
}

func (chnl *Channel) NextRequest(listener Listening, w http.ResponseWriter, r *http.Request, shuttingDownListener func(), shuttingDownHost func(), canShutdownEnv bool) (reqst *Request) {
	chnl.lck.Lock()
	defer chnl.lck.Unlock()
	/*if len(chnl.idlngReqsts) > 0 {
		reqst = chnl.idlngReqsts[0]
		chnl.idlngReqsts = chnl.idlngReqsts[1:]
		reqst.InitRequest(listener, w, r, shuttingDownListener, shuttingDownHost, canShutdownEnv)
	} else {
		reqst = NewRequest(listener, w, r, shuttingDownListener, shuttingDownHost, canShutdownEnv)
	}*/
	reqst = NewRequest(listener, w, r, shuttingDownListener, shuttingDownHost, canShutdownEnv)
	return
}

func (chnl *Channel) EnqueueRequest(reqst *Request) {
	chnl.lck.Lock()
	defer chnl.lck.Unlock()
	reqst.Close()
	//chnl.idlngReqsts = append(chnl.idlngReqsts, reqst)
}
