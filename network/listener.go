package network

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	active "github.com/efjoubert/lnksys/iorw/active"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

/*Listening interface
 */
type Listening interface {
	Shutdown()
	ShutdownHost(string)
	QueueRequest(*Request)
}

type lstnrserver struct {
	httpsvr  *http.Server
	http2svr *http2.Server
	srvmx    *http.ServeMux
}

func newLstnrServer(host string, hdnlr http.Handler,enableh2c bool) (lstnrsvr *lstnrserver) {
	var srvmutex = http.NewServeMux()
	srvmutex.Handle("/", hdnlr)
	var rqsthndlr http.Handler
	var serverh2 = &http2.Server{}
	if enableh2c {
		rqsthndlr=h2c.NewHandler(srvmutex, serverh2)
	} else {
		rqsthndlr:hdnlr
	}
	
	var server = &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		Addr:              host,
		Handler:           rqsthndlr}
	lstnrsvr = &lstnrserver{httpsvr: server, http2svr: serverh2, srvmx: srvmutex}
	return
}

func (lstnrsvr *lstnrserver) listenAndServe() {
	go func(srvr *http.Server) {
		srvr.ListenAndServe()
	}(lstnrsvr.httpsvr)
}

func (lstnrsvr *lstnrserver) Shutdown() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	err = lstnrsvr.httpsvr.Shutdown(ctx)
	lstnrsvr.httpsvr.Close()
	lstnrsvr.httpsvr = nil
	lstnrsvr.http2svr = nil
	lstnrsvr.srvmx = nil
	return
}

/*Listener - Listener
 */
type Listener struct {
	servers        map[string]*lstnrserver
	queuedRequests chan *Request
	qrqstlck       *sync.Mutex
	sema 		   chan struct{}
}

func (lstnr *Listener) QueueRequest(reqst *Request) {
	lstnr.queuedRequests <- reqst
}

func (lstnr *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lstnr.sema <- struct{}{}
	defer func() { <-lstnr.sema }()

	var reqst = NewRequest(lstnr, w, r, func() {
		lstnr.Shutdown()
	}, func() {
		lstnr.ShutdownHost(r.Host)
	}, true)
	reqst.ServeHTTP(w,r)
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
					srv.Shutdown()
					delete(lstnr.servers, host)
				}()
				srv = nil
			}
		}
	}
}

func (lstnr *Listener) ListenAndServer(host string,enableh2c...bool) {
	if host != "" {
		if len(lstnr.servers) == 0 {
			lstnr.servers = map[string]*lstnrserver{}
		}
		if _, hssrv := lstnr.servers[host]; hssrv {
			return
		}
		if len(enableh2c)==1 && enableh2c[0] {
			var server = newLstnrServer(host, lstnr,true)
			server.listenAndServe()
			lstnr.servers[host] = server
		} else {
			var server = newLstnrServer(host, lstnr,false)
			server.listenAndServe()
			lstnr.servers[host] = server
		}
	}
}

var lstnr *Listener

func InvokeListener(host string) {
	lstnr.ListenAndServer(host)
}

func init() {
	if lstnr == nil {
		lstnr = &Listener{sema : make(chan struct{}, runtime.NumCPU()*100), queuedRequests: make(chan *Request, runtime.NumCPU()*101), qrqstlck: &sync.Mutex{}}
		func(qlstnr *Listener) {
			var nmcpus = runtime.NumCPU()
			for nmcpus > 0 {
				nmcpus--
				go func() {
					for {
						select {
						case reqst := <-qlstnr.queuedRequests:
							ExecuteQueuedRequest(reqst)
						}
					}
				}()
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
