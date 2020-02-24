package network

import (
	"net/http"
)

type Channel struct {
}

func NewChannel() (chnl *Channel) {
	chnl = &Channel{}
	return
}

func (chnl *Channel) NextRequest(listener Listening, w http.ResponseWriter, r *http.Request, shuttingDownListener func(), shuttingDownHost func(), canShutdownEnv bool) (reqst *Request) {
	reqst = NewRequest(listener, w, r, shuttingDownListener, shuttingDownHost, canShutdownEnv)
	return
}
