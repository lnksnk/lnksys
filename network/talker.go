package network

import (
	"io"
	"net"
	http "net/http"
	"time"

	iorw "github.com/efjoubert/lnksys/iorw"
	"github.com/efjoubert/lnksys/iorw/active"
	"github.com/efjoubert/lnksys/parameters"
)

type Talking interface {
	Send(url string, body io.Reader, headers map[string][]string, params ...interface{}) (err error)
	FSend(w io.Writer, body io.Reader, headers map[string][]string, url string, params ...interface{}) (err error)
}

//http2 "golang.org/x/net/http2"

type Talker struct {
	client      *http.Client
	trw         *iorw.BufferedRW
	atv         *active.Active
	prms        *parameters.Parameters
	enableClose bool
	h2c         bool
}

const maxBufferSize int64 = 81920

func NewTalker(h2c ...bool) (tlkr *Talker) {
	var netTransport *http.Transport = nil
	//if len(h2c) == 1 && h2c[0] {
	//	netTransport = http2.Transport{}
	//} else {
	netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	//}
	var trwref *iorw.BufferedRW = iorw.NewBufferedRW(maxBufferSize, nil)
	tlkr = &Talker{enableClose: false, trw: trwref, client: &http.Client{Timeout: time.Second * 10, Transport: netTransport},
		prms: parameters.NewParameters(), h2c: len(h2c) == 1 && h2c[0]}
	tlkr.atv = active.NewActive(int64(81920), map[string]interface{}{"out": tlkr.trw})
	return
}

func (tlkr *Talker) Send(url string, body io.Reader, headers map[string][]string, params ...interface{}) (err error) {
	return tlkr.FSend(nil, body, headers, url, params...)
}

func (tlkr *Talker) FSend(w io.Writer, body io.Reader, headers map[string][]string, url string, params ...interface{}) (err error) {
	defer func() {
		tlkr.enableClose = true
	}()
	tlkr.enableClose = false
	var method = "GET"
	if len(params) > 0 {
		method = "POST"
	}
	var req, reqerr = http.NewRequest(method, url, body)
	if len(headers) > 0 {
		for hdrkey, hdrvals := range headers {
			for _, hdrv := range hdrvals {
				req.Header.Add(http.CanonicalHeaderKey(hdrkey), hdrv)
			}
		}
	}
	if reqerr == nil && req != nil {
		var resp, resperr = tlkr.client.Do(req)
		if resperr == nil {
			if resp.Body != nil {
				if w == nil {
					io.Copy(tlkr, resp.Body)
					tlkr.trw.Print(resp.Body)
				} else {
					iorw.FPrint(w, resp.Body)
				}
			}
		}
	} else {
		err = reqerr
	}
	return
}

func (tlkr *Talker) Reset() {
	if tlkr.atv != nil {
		tlkr.atv.Reset()
	}
	if tlkr.trw != nil {
		tlkr.trw.Reset()
	}
}

/*Print iorw.Printing Print
 */
func (tlkr *Talker) Print(a ...interface{}) {
	if tlkr.trw != nil {
		tlkr.trw.Print(a...)
	}
}

/*Println iorw.Printing Print
 */
func (tlkr *Talker) Println(a ...interface{}) {
	if tlkr.trw != nil {
		tlkr.trw.Println(a...)
	}
}

/*Read io.Reader Read
 */
func (tlkr *Talker) Read(p []byte) (n int, err error) {
	n, err = tlkr.trw.Read(p)
	return
}

/*Write io.Writer Write
 */
func (tlkr *Talker) Write(p []byte) (n int, err error) {
	n, err = tlkr.trw.Write(p)
	return
}

func (tlkr *Talker) ReadRune() (r rune, size int, err error) {
	return
}

func (tlkr *Talker) Close() (err error) {
	if tlkr.enableClose {
		if tlkr.trw != nil {
			tlkr.trw.Close()
			tlkr.trw = nil
		}
		if tlkr.atv != nil {
			tlkr.atv.Close()
			tlkr.atv = nil
		}
	}
	return
}
