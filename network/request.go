package network

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	db "github.com/efjoubert/lnksys/db"
	iorw "github.com/efjoubert/lnksys/iorw"
	active "github.com/efjoubert/lnksys/iorw/active"

	mime "github.com/efjoubert/lnksys/network/mime"
	parameters "github.com/efjoubert/lnksys/parameters"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *gzipResponseWriter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	return
}

func (w *gzipResponseWriter) Flush() {
	err := error(nil)
	if gzw, gzwok := w.Writer.(*gzip.Writer); gzwok {
		err = gzw.Flush()
	}
	if err == nil {
		if f, fok := w.ResponseWriter.(http.Flusher); fok {
			f.Flush()
		}
	}
}

const maxbufsize int = 81920

type Request struct {
	rqstlck               *sync.Mutex
	readFromOffset        int64
	readToOffset          int64
	rspnshdrs             map[string]string
	bufRW                 *iorw.BufferedRW
	rw                    *iorw.RW
	rqstContent           *iorw.BufferedRW
	listener              Listening
	talker                Talking
	w                     http.ResponseWriter
	wpipeR                *io.PipeReader
	wpipeW                *io.PipeWriter
	wpipeE                chan error
	r                     *http.Request
	done                  chan bool
	resourcesOffset       int64
	resourcesSize         int64
	currdr                *Resource
	resources             []*Resource
	lastResourcePathAdded string
	isfirstResource       bool
	resourcepaths         []string
	rootpaths             []string
	preCurrentBytes       []byte
	preCurrentBytesl      int
	preCurrentBytesi      int
	currentbytes          []byte
	currentbytesl         int
	currentbytesi         int
	currentrunes          []rune
	currentrunesl         int
	currentrunesi         int
	preCurrentRunes       []byte
	preCurrentRunesl      int
	preCurrentRunesi      int
	runeRdr               *bufio.Reader
	dbcn                  map[string]*db.DbConnection
	params                *parameters.Parameters
	*active.Active
	interuptRequest      bool
	runeBuffer           []rune
	runeBufferErr        error
	runeBuffers          []int
	runeBufferi          int
	runeBufferl          int
	runeBuffering        chan bool
	shuttingdownHost     func()
	canShutdownHost      bool
	shuttingdownListener func()
	canShutdownListener  bool
	shuttingdownEnv      func()
	canShutdownEnv       bool
	forceRead            bool
	busyForcing          bool
	preWriteHeader       func()
}

func (reqst *Request) ReadFromOffset() int64 {
	return reqst.readFromOffset
}

func (reqst *Request) ReadToOffset() int64 {
	return reqst.readToOffset
}

func (reqst *Request) RequestHeader() http.Header {
	return reqst.r.Header
}

func (reqst *Request) ResponseHeader() http.Header {
	return reqst.w.Header()
}

func (reqst *Request) RequestHeaders() (hdrs []string) {
	for h := range reqst.r.Header {
		hdrs = append(hdrs, h)
	}
	return
}

func (reqst *Request) ResponseHeaders() (hdrs []string) {
	for h := range reqst.w.Header() {
		hdrs = append(hdrs, h)
	}
	return
}

func (reqst *Request) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			rqsterr := fmt.Errorf("Panic: %+v\n", err)
			fmt.Println(rqsterr)
		}
		reqst.rqstlck.Unlock()
		reqst.Close()
	}()
	reqst.rqstlck.Lock()
	var wi interface{} = w
	if _, wiok := wi.(*Response); !wiok {
		go func(wnotify <-chan bool, rcntx context.Context) {
			var checking = true
			for checking {
				select {
				case <-wnotify:
					reqst.interuptRequest = true
					checking = false
				case <-rcntx.Done():
					reqst.interuptRequest = true
					checking = false
				}
			}
			return
		}(w.(http.CloseNotifier).CloseNotify(), r.Context())
	}
	reqst.ExecuteRequest()
}

func (reqst *Request) Interupted() bool {
	return reqst.interuptRequest
}

func HttpRequestHandler(reqst *Request) (hndlr http.Handler) {
	hndlr = reqst
	return
}

func (reqst *Request) IsActiveContent(ext string) (active bool) {
	ext = filepath.Ext(ext)
	if _, active = atvExtns[ext]; active {
		active = atvExtns[ext]
	}
	return
}

func (reqst *Request) Db(alias string) (dbcn *db.DbConnection) {
	if reqst.dbcn[alias] == nil {
		if dbcn = db.DBMSManager().Dbms(alias); dbcn != nil {
			reqst.dbcn[alias] = dbcn
		}
	} else {
		dbcn = reqst.dbcn[alias]
	}
	return
}

func (reqst *Request) DbQuery(alias string, query string, args ...interface{}) (dbquery *db.DBQuery) {
	dbquery = db.DBMSManager().Query(alias, query, args...)
	return
}

func (reqst *Request) AddResource(resource ...string) {
	if len(resource) > 0 {
		var lastrsri = len(reqst.resourcepaths)
		var resi = 0
		for len(resource) > 0 {
			var res = resource[0]
			resi = 0
			resource = resource[1:]
			if res != "" {
				if strings.Index(res, "|") > 0 {
					for strings.Index(res, "|") > 0 {
						var rs = res[:strings.Index(res, "|")]
						res = res[strings.Index(res, "|")+1:]
						if rs != "" {
							resource = append(append(resource[:resi], rs), resource[resi:]...)
							resi++
						}
					}
					if res != "" {
						resource = append(append(resource[:resi], res), resource[resi:]...)
					}
				} else {
					if len(reqst.resourcepaths) == 0 {
						reqst.resourcepaths = append(reqst.resourcepaths, res)
					} else {
						reqst.resourcepaths = append(append(reqst.resourcepaths[:lastrsri], res), reqst.resourcepaths[lastrsri:]...)
					}
					lastrsri++
				}
			}
		}
	}
	return
}

func nextResource(reqst *Request, nxtrspath string) (nxtrs *Resource) {
	if nxtrspath != "" {
		nxtrs = NewResource(reqst, nxtrspath)
	}
	return nxtrs
}

func (reqst *Request) RequestContent() *iorw.BufferedRW {
	return reqst.rqstContent
}

func (reqst *Request) ExecuteRequest() {
	var gzw *gzipResponseWriter
	var curResource *Resource = nil
	var isAtv = reqst.IsActiveContent(reqst.r.URL.Path)
	if reqst.bufRW == nil {
		reqst.bufRW = iorw.NewBufferedRW(int64(maxbufsize), reqst)
	}

	var disableActive = false
	var isMultiMedia = false
	var acceptedencoding = reqst.RequestHeader().Get("Accept-Encoding")
	reqst.PopulateParameters()
	if reqst.params.ContainsParameter("disable-active") {
		if disableAtv := reqst.params.Parameter("disable-active"); len(disableAtv) == 1 && strings.ToUpper(disableAtv[0]) == "Y" {
			disableActive = true
		}
	}
	//}
	if isAtv {
		if reqst.rqstContent == nil {
			if reqst.r.Body != nil {
				reqst.rqstContent = iorw.NewBufferedRW(int64(maxbufsize), reqst.r.Body)
			} else {
				reqst.rqstContent = iorw.NewBufferedRW(int64(maxbufsize), nil)
			}
		}
		reqst.forceRead = isAtv
	}
	var mimedetails = mime.FindMimeTypeByExt(reqst.r.URL.Path, ".txt", "text/plain")
	if len(mimedetails) > 0 && strings.HasPrefix(mimedetails[0], "video/") || strings.HasPrefix(mimedetails[0], "audio/") {
		isMultiMedia = true
	}
	var contentencoding = ""

	reqst.w.Header().Set("Cache-Control", "no-store")
	reqst.AddResource(reqst.r.URL.Path)
	if isAtv {
		contentencoding = "; charset=UTF-8"
	}

	if reqst.preWriteHeader == nil {
		defer func() {
			if reqst.preWriteHeader != nil {
				reqst.preWriteHeader()
				reqst.preWriteHeader = nil
			}
		}()
		reqst.preWriteHeader = func() {
			var statusCode = http.StatusOK
			reqst.ResponseHeader().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
			if reqst.ResponseHeader().Get("Content-Type") == "" {
				reqst.ResponseHeader().Set("Content-Type", mimedetails[0]+contentencoding)
			}

			if isMultiMedia {
				/*acceptedranges := "bytes"
				if rangeval := reqst.RequestHeader().Get("Range"); rangeval != "" {
					if strings.Index(rangeval, "=") > 0 {
						acceptedranges = strings.TrimSpace(rangeval[:strings.Index(rangeval, "=")])
						rangeval = strings.TrimSpace(rangeval[strings.Index(rangeval, "=")+1:])
						rangeSize := int64(0)
						if strings.Index(rangeval, "-") > 0 {
							if calcOffset, calcOffSetErr := strconv.ParseInt(rangeval[:strings.Index(rangeval, "-")], 10, 64); calcOffSetErr == nil {
								rangeval = strings.TrimSpace(rangeval[strings.Index(rangeval, "-")+1:])
								rangeSize = curResource.Size()
								if rangeval == "" {
									reqst.readFromOffset = calcOffset
									reqst.readToOffset = rangeSize
								} else if nextcalcOffset, nextcalcOffSetErr := strconv.ParseInt(rangeval, 10, 64); nextcalcOffSetErr == nil {
									reqst.readFromOffset = calcOffset
									reqst.readToOffset = nextcalcOffset
								}
							}
						}
						if reqst.readFromOffset < reqst.readToOffset {
							reqst.ResponseHeader().Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", reqst.readFromOffset, reqst.readToOffset-1, rangeSize))
							reqst.ResponseHeader().Add("Content-Length", fmt.Sprintf("%d", reqst.readToOffset-reqst.readFromOffset))
							statusCode = http.StatusPartialContent
							curResource.Seek(reqst.readFromOffset, 0)
						}
					}
				}
				reqst.ResponseHeader().Set("Content-Encoding", "identity")
				acceptedencoding = "identity"
				reqst.ResponseHeader().Set("Accept-Ranges", acceptedranges)*/
			}
			/*if strings.Index(acceptedencoding, "gzip") >= 0 {
				reqst.ResponseHeader().Set("Content-Encoding", "gzip")
				gzw = &gzipResponseWriter{ResponseWriter: reqst.w, Writer: gzip.NewWriter(reqst.w)}
				reqst.w = gzw
			}*/
			reqst.w.WriteHeader(statusCode)
		}
	}
	if reqst.Active == nil {
		reqst.Active = active.NewActive(int64(maxbufsize), reqst, map[string]interface{}{"DBMS": db.DBMSManager, "Parameters": func() *parameters.Parameters {
			return reqst.Parameters()
		}, "DBQuery": func(alias string, query string, args ...interface{}) (dbquery *db.DBQuery) {
			dbquery = reqst.DbQuery(alias, query, args)
			return
		}, "request": reqst, "SHUTDOWNENV": func() {
			if reqst.shuttingdownEnv != nil {
				reqst.canShutdownEnv = true
			}
		}, "SHUTDOWNHOST": func() {
			if reqst.shuttingdownHost != nil {
				reqst.canShutdownHost = true
			}
		}, "SHUTDOWNLISTENER": func() {
			if reqst.shuttingdownListener != nil {
				reqst.canShutdownListener = true
			}
		}})
	} else {
		reqst.Active.Reset()
	}
	var isFirtsRS = true
	for {
		if len(reqst.resourcepaths) > 0 {
			var nextrs = reqst.resourcepaths[0]
			reqst.resourcepaths = reqst.resourcepaths[1:]
			if !strings.HasPrefix(nextrs, "/") {
				if reqst.lastResourcePathAdded == "" {
					reqst.lastResourcePathAdded = "/"
				}
				nextrs = reqst.lastResourcePathAdded + nextrs
			} else if strings.HasPrefix(nextrs, "/") {
				reqst.lastResourcePathAdded = nextrs[:strings.LastIndex(nextrs, "/")+1]
			}
			if nxtrs := nextResource(reqst, nextrs); nxtrs != nil {
				if curResource == nil || curResource != nxtrs {
					curResource = nxtrs
				}
				if isFirtsRS {
					if !isAtv {
						if !isMultiMedia {
							reqst.ResponseHeader().Set("Content-Length", fmt.Sprintf("%d", nxtrs.Size()))
						}
					}
					isFirtsRS = false
				}
				if isAtv {
					if atverr := func(nxtrs *Resource) (fnerr error) {
						defer func() {
							nxtrs.Close()
						}()
						if nxtrs.disableActive || disableActive {
							iorw.PipedFPrint(reqst, nxtrs)
						} else {
							if nxtrs.activeInverse {
								fnerr = reqst.Active.ACommit("<@", nxtrs, "@>")
							} else {
								if isMultiMedia {
									if reqst.preWriteHeader != nil {
										reqst.preWriteHeader()
										reqst.preWriteHeader = nil
									}
									http.ServeContent(reqst.w, reqst.r, "", time.Now(), nxtrs)
								} else {
									fnerr = reqst.Active.ACommit(nxtrs)
								}
							}
						}
						return
					}(nxtrs); atverr != nil {
						fmt.Print(atverr)
						break
					}
				} else {
					if isMultiMedia {
						if reqst.preWriteHeader != nil {
							reqst.preWriteHeader()
							reqst.preWriteHeader = nil
						}
						http.ServeContent(reqst.w, reqst.r, reqst.r.URL.Path, time.Now(), nxtrs)
						/*fmt.Println()
						fmt.Println("REQUEST-HEADERS")
						for _, hdr := range reqst.RequestHeaders() {
							fmt.Printf("%s:%s\r\n", hdr, reqst.r.Header.Get(hdr))
						}
						fmt.Println()*/
						//io.CopyN(reqst, nxtrs, nxtrs.Size())
						iorw.PipedFPrint(reqst, nxtrs)

						/*fmt.Println("RESPONSE-HEADERS")
						for _, hdr := range reqst.ResponseHeaders() {
							fmt.Printf("%s:%s\r\n", hdr, reqst.w.Header().Get(hdr))
						}*/
					} else {
						iorw.PipedFPrint(reqst, nxtrs)
						//reqst.Print(nxtrs)
					}
				}
			}
		} else {
			break
		}
	}
	if gzw != nil {
		gzw.Flush()
	}
}

func (reqst *Request) Size() int64 {
	return reqst.resourcesSize
}

func (reqst *Request) Parameters() (params *parameters.Parameters) {
	params = reqst.params
	return
}

func (reqst *Request) Seek(offset int64, whence int) (n int64, err error) {
	if whence == io.SeekStart {
		if offset < reqst.Size() {
			n = offset
		}
	} else if whence == io.SeekCurrent {

	} else if whence == io.SeekEnd {
		if offset >= 0 && (reqst.Size()-offset) <= reqst.Size() {
			n = reqst.Size() - offset
		}
	}
	if err == nil {
		reqst.resourcesOffset = n
	}
	return
}

func (reqst *Request) Println(a ...interface{}) {
	reqst.Print(a...)
	reqst.Print("\r\n")
}

func (reqst *Request) Print(a ...interface{}) {
	iorw.FPrint(reqst, a...)
}

func (reqst *Request) ReadRune() (r rune, size int, err error) {
	if reqst.runeBufferl == 0 || (reqst.runeBufferl > 0 && reqst.runeBufferi == reqst.runeBufferl) {
		if reqst.runeBufferErr != nil {
			err = reqst.runeBufferErr
			return
		}
		if reqst.runeBufferi > 0 {
			reqst.runeBufferi = 0
		}
		reqst.runeBufferErr = nil
		reqst.runeBufferl = 0
		if reqst.runeBuffering == nil {
			reqst.runeBuffering = make(chan bool, 1)
		}
		go func() {
			for {
				if reqst.runeRdr == nil {
					reqst.runeRdr = bufio.NewReader(reqst)
				}
				if rr, rsize, rerr := reqst.runeRdr.ReadRune(); rerr == nil {
					if rsize > 0 {
						if len(reqst.runeBuffer) == 0 {
							reqst.runeBuffer = make([]rune, maxbufsize)
						}
						if len(reqst.runeBuffers) == 0 {
							reqst.runeBuffers = make([]int, maxbufsize)
						}
						reqst.runeBuffer[reqst.runeBufferi] = rr
						reqst.runeBuffers[reqst.runeBufferi] = rsize
						reqst.runeBufferi++
						reqst.runeBufferl++
						if len(reqst.runeBuffer) == reqst.runeBufferi {
							break
						}
					}
				} else {
					reqst.runeBufferErr = rerr
					break
				}
			}
			reqst.runeBuffering <- true
		}()
		<-reqst.runeBuffering
		reqst.runeBufferi = 0
	}
	if reqst.runeBufferl > 0 {
		r = reqst.runeBuffer[reqst.runeBufferi]
		size = reqst.runeBuffers[reqst.runeBufferi]
		reqst.runeBufferi++
	} else {
		err = io.EOF
	}
	return
}

func (reqst *Request) WriteTo(w io.Writer) (n int64, err error) {
	var p = make([]byte, maxbufsize)
	var f http.Flusher = nil
	var fok = false
	for {
		if pn, pnerr := reqst.Read(p); pn > 0 || pnerr != nil {
			if pn > 0 {
				var pnn = 0
				for pnn < pn {
					if wn, wnerr := w.Write(p[pnn : pnn+(pn-pnn)]); wn > 0 || wnerr != nil {
						if f == nil {
							if f, fok = reqst.w.(http.Flusher); fok {
								f.Flush()
							}
						}
						pnn += wn
						if wnerr != nil {
							pnerr = wnerr
							break
						}
					}
				}
				n += int64(pnn)
			}
			if pnerr != nil {
				err = pnerr
				break
			}
		}
	}
	p = nil
	return
}

func readingRequest(reqst *Request, p []byte) (n int, err error) {
	if len(reqst.currentbytes) == 0 {
		reqst.currentbytes = make([]byte, maxbufsize)
	}
	var pl = len(p)
	for n < pl && !reqst.Interupted() {
		if (reqst.currentbytesl == 0) || (reqst.currentbytesl > 0 && reqst.currentbytesl == reqst.currentbytesi) {
			if reqst.currentbytesi > 0 {
				reqst.currentbytesi = 0
			}
			if reqst.currentbytesl, err = readResources(reqst, reqst.currentbytes); reqst.currentbytesl == 0 || reqst.Interupted() {
				if err == nil {
					if reqst.currdr == nil {
						err = io.EOF
					} else {
						continue
					}
				}
				break
			}
		}
		for n < pl && reqst.currentbytesi < reqst.currentbytesl && !reqst.Interupted() {
			if (pl - n) >= (reqst.currentbytesl - reqst.currentbytesi) {
				var cl = copy(p[n:n+(reqst.currentbytesl-reqst.currentbytesi)], reqst.currentbytes[reqst.currentbytesi:reqst.currentbytesi+(reqst.currentbytesl-reqst.currentbytesi)])
				reqst.currentbytesi += cl
				n += cl
			} else if (pl - n) < (reqst.currentbytesl - reqst.currentbytesi) {
				var cl = copy(p[n:n+(pl-n)], reqst.currentbytes[reqst.currentbytesi:reqst.currentbytesi+(pl-n)])
				reqst.currentbytesi += cl
				n += cl
			}
		}
	}
	if reqst.Interupted() {
		err = io.EOF
		n = 0
	}
	return
}

func (reqst *Request) Read(p []byte) (n int, err error) {
	return readingRequest(reqst, p)
}

func readResources(reqst *Request, p []byte) (n int, err error) {
	var pl = len(p)
	if reqst.preCurrentBytesi < reqst.preCurrentBytesl {
		for n < pl && reqst.preCurrentBytesi < reqst.preCurrentBytesl {
			if (pl - n) >= (reqst.preCurrentBytesl - reqst.preCurrentBytesi) {
				var cl = copy(p[n:n+(reqst.preCurrentBytesl-reqst.preCurrentBytesi)], reqst.preCurrentBytes[reqst.preCurrentBytesi:reqst.preCurrentBytesi+(reqst.preCurrentBytesl-reqst.preCurrentBytesi)])
				reqst.preCurrentBytesi += cl
				n += cl
			} else if (pl - n) < (reqst.preCurrentBytesl - reqst.preCurrentBytesi) {
				var cl = copy(p[n:n+(pl-n)], reqst.preCurrentBytes[reqst.preCurrentBytesi:reqst.preCurrentBytesi+(pl-n)])
				reqst.preCurrentBytesi += cl
				n += cl
			}
		}
		if reqst.preCurrentBytesl == reqst.preCurrentBytesi {
			reqst.preCurrentBytes = nil
			reqst.preCurrentBytesi = 0
			reqst.preCurrentBytesl = 0
		}
		return
	}
	if reqst.currdr == nil {
		if len(reqst.resources) > 0 {
			reqst.currdr = reqst.resources[0]
			if len(reqst.resources) > 0 {
				reqst.resources = reqst.resources[1:]
			}
		}
		if reqst.currdr == nil {
			err = io.EOF
		} else if reqst.currdr != nil {
			if reqst.currdr.activeInverse {
				n = copy(p, []byte("<@"))
			}
		}
	}
	if reqst.currdr != nil {
		if reqst.currdr.activeInverse {
			if reqst.currdr.activeEnd {
				err = io.EOF
			} else {
				nt, errt := reqst.currdr.Read(p[n : n+(pl-n)])
				if errt != nil {
					err = errt
				}
				n += nt
			}
		} else {
			nt, errt := reqst.currdr.Read(p[n : n+(pl-n)])
			if errt != nil {
				err = errt
			}
			n += nt
		}
		if err == io.EOF {
			var currdr io.Reader = reqst.currdr
			if reqst.currdr.IsActiveContent() {
				if reqst.currdr.activeInverse {
					if reqst.currdr.activeEnd {
						if len(reqst.resources) > 0 {
							n = copy(p[n:], []byte("\r\n"))
						}
						reqst.currdr.activeEnd = false
					} else {
						reqst.currdr.activeEnd = true
						n = copy(p[n:], []byte("@>"))
						err = nil
					}
				} else {
					if len(reqst.resources) > 0 {
						n = copy(p[n:], []byte("\r\n"))
					}
				}
			} else {
				if len(reqst.resources) > 0 {
					n = copy(p[n:], []byte("\r\n"))
				}
			}
			if err == io.EOF {
				reqst.currdr = nil
				if rdclose, rdcloseok := currdr.(io.ReadCloser); rdcloseok {
					rdclose.Close()
					rdclose = nil
				}
				if len(reqst.resources) > 0 {
					err = nil
				}
				currdr = nil
			}
		}
	}
	return
}

func (reqst *Request) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		if reqst.preWriteHeader != nil {
			reqst.preWriteHeader()
			reqst.preWriteHeader = nil
		}
		if reqst.wpipeR == nil && reqst.wpipeW == nil {
			/*reqst.wpipeR,reqst.wpipeW=io.Pipe()
			go func(wpipeR *io.PipeReader,wo io.Writer){
				//var setErr=false
				defer func() {
					if rcvr:=recover(); rcvr!=nil {
						//if !setErr {
							//reqst.wpipeE<-fmt.Errorf("Panic: %+v\n", rcvr)
						//}
					}
					wpipeR.Close()
				}()
				npp:=make([]byte,maxbufsize)
				for {
					np,nperr:=wpipeR.Read(npp)
					if np>0 {
						npi:=0
						for npi<np {
							nwp, nwperr := wo.Write(npp[npi:npi+(np-npi)]);
							if nwp > 0 {
								npi+=nwp
							}
							if nwperr!=nil {
								nperr=nwperr
								break
							}
						}
					}
					if nperr!=nil {
						break
					}
				}
			}(reqst.wpipeR,reqst.w)
			*/
		}
		if reqst.wpipeW != nil {
			n, err = reqst.wpipeW.Write(p)
			if err == nil {
				/*if f, ok := reqst.w.(http.Flusher); ok {
					f.Flush()
				}*/
			}
		} else {
			n, err = reqst.w.Write(p)
		}
	}
	return
}

func NewRequest(listener Listening, w http.ResponseWriter, r *http.Request, shuttingDownListener func(), shuttingDownHost func(), canShutdownEnv bool) (reqst *Request) {
	reqst = &Request{
		readFromOffset:       -1,
		resourcesOffset:      -1,
		isfirstResource:      true,
		rqstlck:              &sync.Mutex{},
		listener:             listener,
		w:                    w,
		r:                    r,
		done:                 make(chan bool, 1),
		resourcesSize:        0,
		params:               parameters.NewParameters(),
		interuptRequest:      false,
		shuttingdownHost:     shuttingDownHost,
		canShutdownHost:      shuttingDownHost != nil,
		shuttingdownListener: shuttingDownListener,
		forceRead:            false,
		busyForcing:          false,
		rootpaths:            []string{}}
	if len(roots) > 0 {
		for rt, _ := range roots {
			reqst.rootpaths = append(reqst.rootpaths, rt)
		}
	}
	if canShutdownEnv {
		reqst.shuttingdownEnv = func() {
			ShutdownEnv()
		}
	}
	return
}

func (reqst *Request) PopulateParameters() {
	parameters.LoadParametersFromHTTPRequest(reqst.params, reqst.r)
}

var shutdownEnv func()

func RegisterShutdownEnv(shuttingdownEnv func()) {
	if shuttingdownEnv != nil {
		if shutdownEnv == nil {
			shutdownEnv = shuttingdownEnv
		}
	}
}

func ShutdownEnv() {
	if shutdownEnv != nil {
		shutdownEnv()
	}
}

func MapActiveExtension(a ...string) {
	for len(a) > 0 {
		if a[0] != "" {
			a[0] = filepath.Ext(a[0])
			if a[0] != "" && a[0] != "." {
				if hsatvext, atvextok := atvExtns[a[0]]; atvextok && !hsatvext {
					atvExtns[a[0]] = true
				} else {
					atvExtns[a[0]] = true
				}
			} else {
				break
			}
		} else {
			break
		}
		a = a[1:]
	}
}

func init() {
	if atvExtns == nil {
		atvExtns = map[string]bool{}
		MapActiveExtension(strings.Split(".html,.htm,.xml,.svg,.css,.js,.json,.txt,.csv", ",")...)
	}

	active.MapGlobals("MAPRoots", func(a ...string) {
		MapRoots(a...)
	})

	active.MapGlobals("MAPActiveExtensions", func(a ...string) {
		MapActiveExtension(a...)
	})
}

func (reqst *Request) Close() (err error) {
	if reqst.done != nil {
		close(reqst.done)
		reqst.done = nil
	}
	if reqst.listener != nil {
		reqst.listener = nil
	}
	if reqst.w != nil {
		reqst.w = nil
	}
	if reqst.r != nil {
		reqst.r = nil
	}
	if reqst.runeRdr != nil {
		reqst.runeRdr = nil
	}
	if reqst.params != nil {
		reqst.params.CleanupParameters()
		reqst.params = nil
	}
	if reqst.bufRW != nil {
		reqst.bufRW.Close()
		reqst.bufRW = nil
	}
	if reqst.Active != nil {
		reqst.Active.Close()
		reqst.Active = nil
	}
	if len(reqst.runeBuffer) > 0 {
		reqst.runeBuffer = nil
		reqst.runeBuffers = nil
		reqst.runeBufferErr = nil
	}
	if reqst.runeBuffering != nil {
		close(reqst.runeBuffering)
		reqst.runeBuffering = nil
	}
	if reqst.shuttingdownHost != nil {
		if !reqst.interuptRequest && reqst.canShutdownHost {
			reqst.shuttingdownHost()
		}
		reqst.shuttingdownHost = nil
	}
	if reqst.shuttingdownListener != nil {
		if !reqst.interuptRequest && reqst.canShutdownListener {
			reqst.shuttingdownListener()
		}
		reqst.shuttingdownHost = nil
	}
	if reqst.shuttingdownEnv != nil {
		if !reqst.interuptRequest && reqst.canShutdownEnv {
			reqst.shuttingdownEnv()
		}
		reqst.shuttingdownEnv = nil
	}
	if reqst.resources != nil {
		for len(reqst.resources) > 0 {
			reqst.resources[0].Close()
			reqst.resources[0] = nil
			reqst.resources = reqst.resources[1:]
		}
		reqst.resources = nil
	}
	if reqst.rqstContent != nil {
		reqst.rqstContent.Close()
		reqst.rqstContent = nil
	}
	if reqst.rootpaths != nil {
		reqst.rootpaths = nil
	}
	if reqst.wpipeR != nil {
		reqst.wpipeR.Close()
		reqst.wpipeR = nil
	}
	if reqst.wpipeW != nil {
		reqst.wpipeW.Close()
		reqst.wpipeW = nil
	}
	if reqst.wpipeE != nil {
		close(reqst.wpipeE)
		reqst.wpipeE = nil
	}
	return
}

var roots = map[string]string{}

func MapRoot(path string, rootpath string) {
	roots[path] = rootpath
}

func MapRoots(rootsToMap ...string) {
	for len(rootsToMap) > 0 && len(rootsToMap)%2 == 0 {
		roots[rootsToMap[0]] = rootsToMap[1]
		if len(rootsToMap) > 2 {
			rootsToMap = rootsToMap[2:]
		} else {
			break
		}
	}
}

var atvExtns map[string]bool

type Response struct {
	r          *http.Request
	w          io.Writer
	statusCode int
	header     http.Header
}

func NewResponse(w io.Writer, r *http.Request) (resp *Response) {
	resp = &Response{w: w, header: http.Header{}, r: r}
	return resp
}

func (resp *Response) Header() http.Header {
	return resp.header
}

func (resp *Response) Write(p []byte) (n int, err error) {
	if resp.w != nil {
		n, err = resp.w.Write(p)
	}
	return 0, nil
}

func (resp *Response) WriteHeader(statusCode int) {
	resp.statusCode = statusCode

	if resp.w != nil {
		var statusLine = resp.r.Proto + " " + fmt.Sprintf("%d", statusCode) + " " + http.StatusText(statusCode)
		fmt.Fprintln(resp.w, statusLine)
		if resp.header != nil {
			resp.header.Write(resp.w)
		}
		fmt.Fprintln(resp.w)
	}
}

func DefaultServeHttp(w io.Writer, method string, url string, body io.Reader) {
	if rhttp, rhttperr := http.NewRequest(method, url, body); rhttperr == nil {
		if rhttp != nil {
			var whttp = NewResponse(w, rhttp)
			var reqst = NewRequest(nil, whttp, rhttp, nil, nil, false)
			HttpRequestHandler(reqst).ServeHTTP(whttp, rhttp)
		}
	}
}

func BrokerServeHttp(w io.Writer, body io.Reader, exename string, exealias string, args ...string) {
	var url = "/"
	var method = "GET"

	if rhttp, rhttperr := http.NewRequest(method, url, body); rhttperr == nil {
		if rhttp != nil {
			var whttp = NewResponse(w, rhttp)
			var reqst = NewRequest(nil, whttp, rhttp, nil, nil, false)
			HttpRequestHandler(reqst).ServeHTTP(whttp, rhttp)
		}
	}
}

var callableResources map[string]func() io.Reader

func RegisterCallableResource(resource string, callable func() io.Reader, a ...interface{}) {

}
