package network

import (
	/**/
	"context"
	_ "encoding/gob"
	"net/http"
	_ "net/rpc"
	_ "reflect"
	"time"

	active "github/efjoubert/lnksys/iorw/active"
)

/*RPCRequest RPCRequest
*/
type RPCRequest struct {
}

/*RPCListener - RPCListener
 */
type RPCListener struct {
	servers     map[string]*http.Server
	servmutexes map[string]*http.ServeMux
}

func (rpclstnr *RPCListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var reqst = NewRequest(rpclstnr, w, r, func() {
		rpclstnr.Shutdown()
	}, func() {
		rpclstnr.ShutdownHost(r.Host)
	}, true)
	HttpRequestHandler(reqst).ServeHTTP(w, r)
}

func (rpclstnr *RPCListener) Shutdown() {
	if len(rpclstnr.servers) > 0 {
		var hosts = []string{}
		for host, _ := range rpclstnr.servers {
			hosts = append(hosts, host)
		}
		for _, host := range hosts {
			rpclstnr.ShutdownHost(host)
		}
	}
}

func (rpclstnr *RPCListener) ShutdownHost(host string) {
	if host != "" {
		if len(rpclstnr.servers) > 0 {
			if srv, srvok := lstnr.servers[host]; srvok {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer func() {
					cancel()
				}()
				if err := srv.Shutdown(ctx); err != nil {

				}
				delete(rpclstnr.servers, host)
				rpclstnr.servmutexes[host] = nil
				delete(rpclstnr.servmutexes, host)
				srv.Close()
				srv = nil
			}
		}
	}
}

func (rpclstnr *RPCListener) ListenAndServer(host string) {
	if host != "" {
		if len(rpclstnr.servers) == 0 {
			rpclstnr.servers = map[string]*http.Server{}
		}
		if len(rpclstnr.servmutexes) == 0 {
			rpclstnr.servmutexes = map[string]*http.ServeMux{}
		}
		if _, hssrv := rpclstnr.servers[host]; hssrv {
			return
		}
		var srvmutex = http.NewServeMux()
		srvmutex.Handle("/", lstnr)
		var server = &http.Server{Addr: host, Handler: srvmutex}
		rpclstnr.servers[host] = server
		go func() {
			server.ListenAndServe()
		}()
	}
}

var rpclstnr *RPCListener

func InvokeRPCListener(host string) {
	rpclstnr.ListenAndServer(host)
}

func init() {
	if rpclstnr == nil {
		rpclstnr = &RPCListener{}
	}
	active.MapGlobals("InvokeRPCListener", InvokeRPCListener)
}

func ShutdownRPCListener() {
	if rpclstnr != nil {
		rpclstnr.Shutdown()
	}
}
