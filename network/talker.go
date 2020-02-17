package network

import (
	"bufio"
	"encoding/json"
	"io"
	"mime/multipart"
	"net"
	http "net/http"
	"strings"
	"time"

	iorw "github.com/efjoubert/lnksys/iorw"
	"github.com/efjoubert/lnksys/iorw/active"
	mime "github.com/efjoubert/lnksys/network/mime"
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
	if headers == nil {
		headers = map[string][]string{}
	}
	var urlPath = url
	if strings.Index(urlPath, "?") > -1 {
		urlPath = urlPath[:strings.Index(urlPath, "?")]
	}
	var mimedetails = mime.FindMimeTypeByExt(urlPath, ".txt", "text/plain")
	var mimetype = mimedetails[0]
	headers["Content-Type"] = append(headers["Content-Type"], mimetype)

	if len(params) > 0 {
		pipeReader, pipeWriter := io.Pipe()
		bufPipeW := bufio.NewWriter(pipeWriter)
		ispmap := false
		var pmap map[string]interface{}
		for _, pchk := range params {
			if pmpi, pmpiok := pchk.(map[string]interface{}); pmpiok {
				if ispmap = len(pmpi) > 0; ispmap {
					pmap = pmpi
				}
				break
			}
		}

		method = "POST"
		if ispmap {
			enc := json.NewEncoder(bufPipeW)
			headers["Content-Type"] = []string{"application/json"}
			go func() {
				defer func() {
					pipeWriter.Close()
				}()
				rqstprms := map[string][]string{}
				for _, d := range params {
					if prms, prmsok := d.(*parameters.Parameters); prmsok {
						for _, prmstd := range prms.StandardKeys() {
							if prmstdval := prms.Parameter(prmstd); len(prmstdval) > 0 {
								if rqstv, rqstkok := rqstprms[prmstd]; rqstkok {
									rqstprms[prmstd] = append(rqstv, prmstdval...)
								} else {
									rqstprms[prmstd] = prmstdval[:]
								}
							}
						}
					} else if prms, prmsok := d.(map[string]string); prmsok {
						for pk, pv := range prms {
							if rqstv, rqstkok := rqstprms[pk]; rqstkok {
								rqstprms[pk] = append(rqstv, pv)
							} else {
								rqstprms[pk] = []string{pv}
							}
						}
					} else if prms, prmsok := d.(map[string][]string); prmsok {
						for pk, pv := range prms {
							if len(pv) > 0 {
								if rqstv, rqstkok := rqstprms[pk]; rqstkok {
									rqstprms[pk] = append(rqstv, pv...)
								} else {
									rqstprms[pk] = pv[:]
								}
							}
						}
					}
				}
				if rqstprmsfound, rqstprmsok := pmap["reqst-params"]; rqstprmsok {
					if rqstprmsmp, rqstpmsmpok := rqstprmsfound.(map[string]interface{}); rqstpmsmpok {
						for pk, pv := range rqstprms {
							var pvi interface{} = pv
							rqstprmsmp[pk] = pvi
						}
					}
				} else {
					pmap["reqst-params"] = rqstprms
				}

				if err = enc.Encode(&pmap); err == nil {
					err = bufPipeW.Flush()
				}
			}()
		} else {
			mpartwriter := multipart.NewWriter(bufPipeW)
			headers["Content-Type"] = []string{mpartwriter.FormDataContentType()}
			go func() {
				defer func() {
					pipeWriter.Close()
				}()
				for _, d := range params {
					if prms, prmsok := d.(*parameters.Parameters); prmsok {
						for _, prmstd := range prms.StandardKeys() {
							for _, prmstdval := range prms.Parameter(prmstd) {
								if part, err := mpartwriter.CreateFormField(prmstd); err != nil {
									break
								} else if _, err = io.Copy(part, strings.NewReader(prmstdval)); err != nil {
									break
								}
							}
							if err != nil {
								break
							}
						}
					} else if prms, prmsok := d.(map[string]string); prmsok {
						for pk, pv := range prms {
							if part, err := mpartwriter.CreateFormField(pk); err != nil {
								break
							} else if _, err = io.Copy(part, strings.NewReader(pv)); err != nil {
								break
							}
						}
					} else if prms, prmsok := d.(map[string][]string); prmsok {
						for pk, pv := range prms {
							for _, pvv := range pv {
								if part, err := mpartwriter.CreateFormField(pk); err != nil {
									break
								} else if _, err = io.Copy(part, strings.NewReader(pvv)); err != nil {
									break
								}
							}
							if err != nil {
								break
							}
						}
					}
					if err != nil {
						break
					}
				}
				if err == nil {
					if err = mpartwriter.Close(); err == nil {
						err = bufPipeW.Flush()
					}
				}
				//errChan <- err
			}()
		}
		body = pipeReader
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
					iorw.PipedFPrint(w, resp.Body)
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
