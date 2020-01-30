package network

import (
	"context"
	"net/http"
	//"runtime"
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
		IdleTimeout:       1 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		Addr:              host,
		Handler:           h2c.NewHandler(hndlr, serverh2)}

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
	tc.SetReadBuffer(4096)
	if err = tc.SetKeepAlive(true); err != nil {
		return nil, err
	}
	// OpenBSD has no user-settable per-socket TCP keepalive
	// https://github.com/caddyserver/caddy/pull/2787
	/*if runtime.GOOS != "openbsd" {
		if err = tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
			return nil, err
		}
	}*/

	return tc, nil
}

func (tcpln tcpKeepAliveListener) File() (*os.File, error) {
	return tcpln.TCPListener.File()
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
	//lstnrsvr.srvmx = nil
	return
}

/*Listener - Listener
 */
type Listener struct {
	servers        map[string]*lstnrserver
	srvlck         *sync.Mutex
}

func (lstnr *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	HttpRequestHandler(func() (rqst *Request) {
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
		lstnr = &Listener{srvlck: &sync.Mutex{}}
	}
	active.MapGlobals("InvokeListener", InvokeListener)
}

func ShutdownListener() {
	if lstnr != nil {
		lstnr.Shutdown()
	}
}
