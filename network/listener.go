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
	"net"
	"os"
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
	//srvmx    *http.ServeMux
	sema chan struct{}
}

func newLstnrServer(host string, hndlr http.Handler) (lstnrsvr *lstnrserver) {

	//var srvmutex = http.NewServeMux()

	//srvmutex.Handle("/", hdnlr)

	var serverh2 = &http2.Server{}

	var server = &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       1 * time.Minute,
		IdleTimeout:       10 * time.Second,
		WriteTimeout:      2 * time.Minute,
		Addr:              host,
		Handler:           h2c.NewHandler(hndlr, serverh2),
		ConnContext: func(ctx context.Context, c net.Conn) (cntx context.Context) {
			cntx = ctx
			return
		}}

	server.SetKeepAlivesEnabled(true)
	lstnrsvr = &lstnrserver{httpsvr: server, http2svr: serverh2}
	return
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (tcpln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := tcpln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetReadBuffer(8192)
	tc.SetWriteBuffer(8192)
	tc.SetNoDelay(false)
	if err = tc.SetKeepAlive(true); err != nil {
		return nil, err
	}
	// OpenBSD has no user-settable per-socket TCP keepalive
	// https://github.com/caddyserver/caddy/pull/2787
	if runtime.GOOS != "openbsd" {
		if err = tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
			return nil, err
		}
	}

	return tc, nil
}

func (tcpln tcpKeepAliveListener) File() (*os.File, error) {
	return tcpln.TCPListener.File()
}

func (lstnrsvr *lstnrserver) listenAndServe() {
	go func(srvr *http.Server) {
		ln, err := net.Listen("tcp", srvr.Addr)
		if err != nil {
			var succeeded bool
			if runtime.GOOS == "windows" {
				// Windows has been known to keep sockets open even after closing the listeners.
				// Tests reveal this error case easily because they call Start() then Stop()
				// in succession. TODO: Better way to handle this? And why limit this to Windows?
				for i := 0; i < 20; i++ {
					time.Sleep(100 * time.Millisecond)
					ln, err = net.Listen("tcp", srvr.Addr)
					if err == nil {
						succeeded = true
						break
					}
				}
			}
			if succeeded {
				if tcpLn, ok := ln.(*net.TCPListener); ok {
					ln = tcpKeepAliveListener{TCPListener: tcpLn}
				}
			}
		}
		if err == nil && ln != nil {
			srvr.Serve(ln)
		}
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
	//lstnrsvr.srvmx = nil
	return
}

/*Listener - Listener
 */
type Listener struct {
	servers        map[string]*lstnrserver
	queuedRequests chan *Request
	qrqstlck       *sync.Mutex
	srvlck         *sync.Mutex
}

func (lstnr *Listener) QueueRequest(reqst *Request) {
	lstnr.queuedRequests <- reqst
}

func (lstnr *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	HttpRequestHandler(func() (rqst *Request) {
		//lstnr.srvlck.Lock()
		//defer func() {
		//	lstnr.srvlck.Unlock()
		//}()
		rqst = NewRequest(lstnr, w, r, func() {
			lstnr.Shutdown()
		}, func() {
			lstnr.ShutdownHost(r.Host)
		}, true)
		return
	}()).ServeHTTP(w, r)
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

func (lstnr *Listener) ListenAndServer(host string) {
	if host != "" {
		if len(lstnr.servers) == 0 {
			lstnr.servers = map[string]*lstnrserver{}
		}
		if _, hssrv := lstnr.servers[host]; hssrv {
			return
		}
		var server = newLstnrServer(host, lstnr)
		server.listenAndServe()
		lstnr.servers[host] = server

	}
}

var lstnr *Listener

func InvokeListener(host string) {
	lstnr.ListenAndServer(host)
}

func init() {
	if lstnr == nil {
		lstnr = &Listener{queuedRequests: make(chan *Request, runtime.NumCPU()*4), qrqstlck: &sync.Mutex{}, srvlck: &sync.Mutex{}}
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
