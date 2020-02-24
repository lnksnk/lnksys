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

func (chnl *Channel) NextRequest(w http.ResponseWriter, r *http.Request, shuttingDownListener func(), shuttingDownHost func(), canShutdownEnv bool) (reqst *Request) {
	reqst = NewRequest(chnl, w, r, shuttingDownListener, shuttingDownHost, canShutdownEnv)
	return
}

func (chnl *Channel) EnqueueRequest(reqst *Request) {
	func() {
		chnl.idlngReqsts = append(chnl.idlngReqsts, reqst)
	}()
}
