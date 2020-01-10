package network

import (
	"context"
	"net/http"
	"sync"
	"time"

	active "github.com/efjoubert/lnksys/iorw/active"
)

/*Listening interface
 */
type Listening interface {
	Shutdown()
	ShutdownHost(string)
	QueueRequest(*Request)
}

/*Listener - Listener
 */
type Listener struct {
	servers        map[string]*http.Server
	servmutexes    map[string]*http.ServeMux
	queuedRequests chan *Request
	qrqstlck       *sync.Mutex
}

func (lstnr *Listener) QueueRequest(reqst *Request) {
	lstnr.qrqstlck.Lock()
	defer lstnr.qrqstlck.Unlock()
	lstnr.queuedRequests <- reqst
}

func (lstnr *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqst = NewRequest(lstnr, w, r, func() {
		lstnr.Shutdown()
	}, func() {
		lstnr.ShutdownHost(r.Host)
	}, true)
	HttpRequestHandler(reqst).ServeHTTP(w, r)
}

func (lstnr *Listener) Shutdown() {
	if len(lstnr.servers) > 0 {
		var hosts = []string{}
		for host, _ := range lstnr.servers {
			hosts = append(hosts, host)
		}
		for _, host := range hosts {
			lstnr.ShutdownHost(host)
		}
	}
}

func (lstnr *Listener) ShutdownHost(host string) {
	if host != "" {
		if len(lstnr.servers) > 0 {
			if srv, srvok := lstnr.servers[host]; srvok {
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer func() {
						cancel()
					}()
					if err := srv.Shutdown(ctx); err != nil {

					}
					delete(lstnr.servers, host)
					lstnr.servmutexes[host] = nil
					delete(lstnr.servmutexes, host)
					srv.Close()
				}()
				srv = nil
			}
		}
	}
}

func (lstnr *Listener) ListenAndServer(host string) {
	if host != "" {
		if len(lstnr.servers) == 0 {
			lstnr.servers = map[string]*http.Server{}
		}
		if len(lstnr.servmutexes) == 0 {
			lstnr.servmutexes = map[string]*http.ServeMux{}
		}
		if _, hssrv := lstnr.servers[host]; hssrv {
			return
		}
		var srvmutex = http.NewServeMux()
		srvmutex.Handle("/", lstnr)
		var server = &http.Server{
			ReadHeaderTimeout:20 *time.Second,
			Addr: host, 
			Handler: srvmutex}
		lstnr.servers[host] = server
		go func(srvr *http.Server) {
			srvr.ListenAndServe()
		}(server)
	}
}

var lstnr *Listener

func InvokeListener(host string) {
	lstnr.ListenAndServer(host)
}

func init() {
	if lstnr == nil {
		lstnr = &Listener{queuedRequests: make(chan *Request), qrqstlck: &sync.Mutex{}}
		go func(qlstnr *Listener) {
			for {
				select {
				case reqst := <-qlstnr.queuedRequests:
					go func() {
						reqst.ExecuteRequest()
						reqst.done <- true
					}()
				}
			}
		}(lstnr)
	}
	active.MapGlobals("InvokeListener", InvokeListener)
}

func ShutdownListener() {
	if lstnr != nil {
		lstnr.Shutdown()
	}
}
