package network

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"bufio"
	"path/filepath"

	db "github.com/efjoubert/lnksys/db"
	embed "github.com/efjoubert/lnksys/embed"
	iorw "github.com/efjoubert/lnksys/iorw"
	active "github.com/efjoubert/lnksys/iorw/active"
	parameters "github.com/efjoubert/lnksys/parameters"
	gzip "github.com/efjoubert/lnksys/network/gzip"
	mime "github.com/efjoubert/lnksys/network/mime"
)

const maxbufsize int = 81920

type Request struct {
	bufRW            *iorw.BufferedRW
	rw               *iorw.RW
	listener         Listening
	w                http.ResponseWriter
	r                *http.Request
	done             chan bool
	resourcesOffset  int64
	resourcesSize    int64
	resdone          chan bool
	currdr           *Resource
	preCurrentBytes  []byte
	preCurrentBytesl int
	preCurrentBytesi int
	currentbytes     []byte
	currentbytesl    int
	currentbytesi    int
	currentrunes     []rune
	currentrunesl    int
	currentrunesi    int
	preCurrentRunes  []byte
	preCurrentRunesl int
	preCurrentRunesi int
	firstrdr         *Resource
	lastrdr          *Resource
	runeRdr          *bufio.Reader
	dbcn             map[string]*db.DbConnection
	params           *parameters.Parameters
	*active.Active
	ctx                  context.Context
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
}

func (reqst *Request) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	reqst.ctx = ctx
	reqstsQueue <- reqst
	var wi interface{} = w
	defer func() {
		var isInterupted = reqst.interuptRequest
		reqst.Close()
		if isInterupted {
			cancel()
		}
	}()
	if _, wiok := wi.(*Response); !wiok {
		go func() {
			var checking = true
			for checking {
				select {
				case <-w.(http.CloseNotifier).CloseNotify():
					reqst.interuptRequest = true
					checking = false
				case <-r.Context().Done():
					reqst.interuptRequest = true
					checking = false
				}
			}
			return
		}()
	}

	<-reqst.done
}

func (reqst *Request) Interupted() bool {
	return reqst.interuptRequest
}

func HttpRequestHandler(reqst *Request) (hndlr http.Handler) {
	if reqst.IsActiveContent(reqst.r.URL.Path) {
		hndlr = gzip.GzipHandler(reqst)
	} else {
		hndlr = reqst
	}
	return
}

func (reqst *Request) IsActiveContent(ext string) (active bool) {
	ext = filepath.Ext(ext)
	active = strings.Index(",.html,.htm,.xml,.svg,.css,.js,.json,", ","+ext+",") > -1
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
	for _, res := range resource {
		if len(res) == 0 {
			continue
		}
		if strings.Index(res, "|") > 0 {
			reqst.AddResource(strings.Split(res, "|")...)
			continue
		}
		if reqst.currdr == nil {
			if rsrc := reqst.NewResource(res, reqst.lastrdr, nil); rsrc != nil {
				reqst.resourcesSize = reqst.resourcesSize + rsrc.size

				if reqst.firstrdr == nil {
					reqst.firstrdr = rsrc
				}
				if reqst.lastrdr == nil {
					reqst.lastrdr = rsrc
				} else {
					reqst.lastrdr.nextrsrc = rsrc
				}
				reqst.lastrdr = rsrc
			}
		} else {
			if rsrc := reqst.NewResource(res, reqst.currdr, reqst.currdr.nextrsrc); rsrc != nil {
				reqst.resourcesSize = reqst.resourcesSize + rsrc.size
				reqst.currdr = rsrc
			}
		}
	}
}

func (reqst *Request) ExecuteRequest() {
	var reqstContentType = reqst.r.Header.Get("Content-Type")
	if reqst.bufRW == nil {
		reqst.bufRW = iorw.NewBufferedRW(int64(maxbufsize), reqst)
	}
	if reqstContentType == "application/json" {

	} else {
		reqst.PopulateParameters()
	}
	var mimedetails = mime.FindMimeTypeByExt(reqst.r.URL.Path, ".txt", "text/plain")
	reqst.AddResource(reqst.r.URL.Path)
	var contentencoding = ""
	reqst.w.Header().Set("Cache-Control", "no-store")
	if reqst.IsActiveContent(reqst.r.URL.Path) {
		contentencoding = "; charset=UTF-8"
		reqst.w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		reqst.w.Header().Set("Content-Type", mimedetails[0]+contentencoding)
		reqst.w.WriteHeader(200)

		if reqst.Active == nil {
			reqst.Active = active.NewActive(maxbufsize,reqst, map[string]interface{}{"DBMS": db.DBMSManager, "Parameters": func() *parameters.Parameters {
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

		if atverr := func() (fnerr error) {
			if fnerr = reqst.Active.APrint(reqst); fnerr==nil {
				fnerr = reqst.Active.ACommit()
			}			
			return
		}(); atverr != nil {
			fmt.Print(atverr)
		}
	} else {
		reqst.w.Header().Set("Content-Type", mimedetails[0]+contentencoding)
		http.ServeContent(reqst.w, reqst.r, reqst.r.URL.Path, time.Now(), reqst.bufRW)
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
	for _, d := range a {
		if r, rok := d.(io.Reader); rok {
			io.Copy(reqst.w, r)
		} else if uarr, uarrok := d.([]uint8); uarrok {
			fmt.Fprint(reqst.w, string(uarr))
		} else if runearr, runearrok := d.([]rune); runearrok {
			fmt.Fprint(reqst.w, string(runearr))
		} else if barr, barrok := d.([]byte); barrok {
			fmt.Fprint(reqst.w, string(barr))
		} else {
			fmt.Fprint(reqst.w, d)
		}
	}
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
	for {
		if pn, pnerr := reqst.Read(p); pn > 0 || pnerr != nil {
			if pn > 0 {
				var pnn = 0
				for pnn < pn {
					if wn, wnerr := w.Write(p[pnn : pnn+(pn-pnn)]); wn > 0 || wnerr != nil {
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

func (reqst *Request) Read(p []byte) (n int, err error) {
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
		go func() {
			var tmpoffset = int64(0)
			var startrdr = reqst.firstrdr
			for tmpoffset <= reqst.resourcesOffset && startrdr != nil {
				if tmpoffset+startrdr.Size() < reqst.resourcesOffset {
					tmpoffset = tmpoffset + startrdr.Size()
					if startrdr.nextrsrc != nil {
						startrdr = startrdr.nextrsrc
					} else {
						break
					}
				} else {
					tmpoffset = reqst.resourcesOffset - tmpoffset
					if sn, snerr := startrdr.Seek(tmpoffset, 0); snerr == nil {
						if sn == tmpoffset {
							reqst.currdr = startrdr
							break
						}
					}
					break
				}
			}
			reqst.resdone <- true
		}()
		<-reqst.resdone
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
			if reqst.currdr.actviveEnd {
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
					if reqst.currdr.actviveEnd {
						if reqst.currdr.nextrsrc != nil {
							n = copy(p[n:], []byte("\r\n"))
						}
						reqst.currdr.actviveEnd = false
					} else {
						reqst.currdr.actviveEnd = true
						n = copy(p[n:], []byte("@>"))
						err = nil
					}
				} else {
					if reqst.currdr.nextrsrc != nil {
						n = copy(p[n:], []byte("\r\n"))
					}
				}
			} else {
				if reqst.currdr.nextrsrc != nil {
					n = copy(p[n:], []byte("\r\n"))
				}
			}
			if err == io.EOF {
				if reqst.currdr.nextrsrc != nil {
					reqst.currdr = reqst.currdr.nextrsrc
					err = nil
				} else {
					reqst.currdr = nil
				}
				if rdclose, rdcloseok := currdr.(io.ReadCloser); rdcloseok {
					rdclose.Close()
					rdclose = nil
				}
				currdr = nil
			}
		}
	}
	return
}

func (reqst *Request) Write(p []byte) (n int, err error) {
	return
}

func NewRequest(listener Listening, w http.ResponseWriter, r *http.Request, shuttingDownListener func(), shuttingDownHost func(), canShutdownEnv bool) (reqst *Request) {
	reqst = &Request{listener: listener, w: w, r: r, done: make(chan bool, 1), resdone: make(chan bool, 1), resourcesSize: 0, params: parameters.NewParameters(), interuptRequest: false,
		shuttingdownHost:     shuttingDownHost,
		canShutdownHost:      shuttingDownHost != nil,
		shuttingdownListener: shuttingDownListener}
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

var reqstsQueue chan *Request

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

func init() {
	if reqstsQueue == nil {
		reqstsQueue = make(chan *Request, runtime.NumCPU()*4)
		go func() {
			for {
				select {
				case reqst := <-reqstsQueue:
					go func() {
						reqst.ExecuteRequest()
						reqst.done <- true
					}()
				}
			}
		}()
	}

	active.MapGlobals("MAPRoots", func(a ...string) {
		MapRoots(a...)
	})
}

func (reqst *Request) Close() (err error) {
	if reqst.done != nil {
		close(reqst.done)
		reqst.done = nil
	}
	if reqst.resdone != nil {
		close(reqst.resdone)
		reqst.resdone = nil
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

type Resource struct {
	reqst         *Request
	finfo         os.FileInfo
	r             io.Reader
	path          string
	pathroot      string
	size          int64
	prvrsrc       *Resource
	nextrsrc      *Resource
	readBuffer    []byte
	readBufferi   int
	readBufferl   int
	rbuf          *bufio.Reader
	activeInverse bool
	actviveEnd    bool
}

func (rsrc *Resource) ReadRune() (r rune, size int, err error) {
	if rsrc.rbuf == nil {
		rsrc.rbuf = bufio.NewReader(rsrc.r)
	}
	r, size, err = rsrc.rbuf.ReadRune()
	return
}
func (rsrc *Resource) IsActiveContent() (active bool) {
	var ext = filepath.Ext(rsrc.path)
	active = strings.Index(",.html,.htm,.xml,.svg,.css,.js,.json,", ","+ext+",") > -1
	return
}

func (reqst *Request) NewResource(resourcepath string, prevrsrc *Resource, nextrsrc *Resource) (rsrc *Resource) {

	var r io.Reader = nil

	var ressplit = strings.Split(resourcepath, "/")
	var tmpres = "/"

	var finfo os.FileInfo = nil
	var lastPathRoot = ""
	var findR = func(rspath string) (rf io.Reader) {
		if rf = embed.EmbedFindJS(rspath); rf == nil {
			for nrs := range ressplit {
				tmpres = strings.Join(ressplit[:nrs+1], "/") + "/"
				if nrs > 0 {
					for root := range roots {
						var zipresource = roots[root] + tmpres[:len(tmpres)-1] + ".zip"
						if _, fiziperr := os.Stat(zipresource); fiziperr == nil {
							func() {
								if zipr, ziprerr := zip.OpenReader(zipresource); ziprerr == nil {
									for _, f := range zipr.File {

										if f.Name == strings.Join(ressplit[nrs+1:], "/") {
											if ziprrc, ziprrcerr := f.Open(); ziprrcerr == nil {
												rf = ziprrc
												finfo = f.FileInfo()
												break
											}
											break
										}
									}
								}
							}()
						} else {
							var resource = roots[root] + tmpres + strings.Join(ressplit[nrs+1:], "/")
							if fi, fierr := os.Stat(resource); fierr == nil {
								if !fi.IsDir() {
									finfo = fi
									lastPathRoot = roots[root] + tmpres
									break
								}
							}
						}
						if rf != nil || finfo != nil {
							break
						}
					}
				} else {
					for root := range roots {
						var resource = roots[root] + tmpres + strings.Join(ressplit[nrs+1:], "/")
						if fi, fierr := os.Stat(resource); fierr == nil {
							if !fi.IsDir() {
								lastPathRoot = roots[root] + tmpres
								finfo = fi
								break
							}
						}
					}
				}
				if rf != nil || finfo != nil {
					break
				}
			}
		}
		return
	}
	var activeInverse = false
	if r = findR(resourcepath); r == nil && finfo == nil && strings.Count(resourcepath, "@") == 2 && strings.Index(resourcepath, "@") > 0 && strings.Index(resourcepath, "@") != strings.LastIndex(resourcepath, "@") {
		activeInverse = true
		resourcepath = strings.Replace(resourcepath, "@", "", -1)
		ressplit = strings.Split(resourcepath, "/")
		r = findR(resourcepath)
	}
	if r != nil || finfo != nil {
		rsrc = &Resource{path: resourcepath, pathroot: lastPathRoot, r: r, finfo: finfo, reqst: reqst, prvrsrc: prevrsrc, nextrsrc: nextrsrc, activeInverse: activeInverse, actviveEnd: false}
		if finfo != nil {
			rsrc.size = finfo.Size()
		}
	}
	return
}

func (rsrc *Resource) Size() int64 {
	return rsrc.size
}

func (rsrc *Resource) ReadRuneBytes(p []byte) (n int, err error) {
	for n < len(p) {
		if c, sz, rerr := rsrc.ReadRune(); rerr != nil {
			if rerr == io.EOF {
				if n == 0 {
					err = rerr
				}
				break
			} else {
				err = rerr
			}
		} else {
			if sz > 0 {
				for _, b := range []byte(string(c)) {
					p[n] = b
					n++
				}
			}
		}
	}
	return
}

func (rsrc *Resource) Read(p []byte) (n int, err error) {
	if rsrc.reqst.interuptRequest {
		err = io.EOF
		return
	}
	var pl = len(p)
	for pl > n {
		if rsrc.readBufferl == 0 || (rsrc.readBufferl > 0 && rsrc.readBufferi == rsrc.readBufferl) {
			if rsrc.readBufferi > 0 {
				rsrc.readBufferi = 0
			}
			if len(rsrc.readBuffer) == 0 {
				rsrc.readBuffer = make([]byte, maxbufsize)
			}
			if rsrc.r == nil && rsrc.finfo != nil {
				if strings.HasSuffix(rsrc.pathroot, "/") && rsrc.pathroot != "/" {
					rsrc.r, _ = os.Open(rsrc.pathroot[:len(rsrc.pathroot)-1] + rsrc.path)
				} else {
					rsrc.r, _ = os.Open(rsrc.pathroot + rsrc.path)
				}
			}
			if rsrc.r != nil {
				if rsrc.IsActiveContent() {
					if rsrc.readBufferl, err = rsrc.ReadRuneBytes(rsrc.readBuffer); err != nil {
						if err == io.EOF {
							if rsrc.readBufferl == 0 {
								rsrc.reqst.resourcesOffset -= rsrc.Size()
								rsrc.reqst.resourcesSize -= rsrc.Size()
								rsrc.readBufferl = 0
								break
							}
						}
					}
				} else {
					if rsrc.readBufferl, err = rsrc.r.Read(rsrc.readBuffer); err != nil {
						if err == io.EOF {
							if rsrc.readBufferl == 0 {
								rsrc.reqst.resourcesOffset -= rsrc.Size()
								rsrc.reqst.resourcesSize -= rsrc.Size()
								rsrc.readBufferl = 0
								break
							}
						}
					}
				}
			} else {
				rsrc.reqst.resourcesOffset -= rsrc.Size()
				rsrc.reqst.resourcesSize -= rsrc.Size()
				err = io.EOF
				rsrc.readBufferl = 0
				break
			}
		}
		for pl > n && rsrc.readBufferl > rsrc.readBufferi {
			if (pl - n) >= (rsrc.readBufferl - rsrc.readBufferi) {
				var rl = copy(p[n:n+(rsrc.readBufferl-rsrc.readBufferi)], rsrc.readBuffer[rsrc.readBufferi:rsrc.readBufferi+(rsrc.readBufferl-rsrc.readBufferi)])
				n += rl
				rsrc.readBufferi += rl
				rsrc.reqst.resourcesOffset += int64(rl)
			} else if (pl - n) < (rsrc.readBufferl - rsrc.readBufferi) {
				var rl = copy(p[n:n+(pl-n)], rsrc.readBuffer[rsrc.readBufferi:rsrc.readBufferi+(pl-n)])
				n += rl
				rsrc.readBufferi += rl
				rsrc.reqst.resourcesOffset += int64(rl)
			}
		}
		if rsrc.readBufferi == rsrc.readBufferl {
			break
		}
	}
	return
}

func (rsrc *Resource) Seek(offset int64, whence int) (n int64, err error) {
	if rsrc.r == nil && rsrc.finfo != nil {
		if strings.HasSuffix(rsrc.pathroot, "/") && rsrc.pathroot != "/" {
			rsrc.r, _ = os.Open(rsrc.pathroot[:len(rsrc.pathroot)-1] + rsrc.path)
		} else {
			rsrc.r, _ = os.Open(rsrc.pathroot + rsrc.path)
		}
	}
	if rs, rsok := rsrc.r.(io.Seeker); rsok {
		n, err = rs.Seek(offset, whence)
	}
	return
}

func (rsrc *Resource) Close() (err error) {
	if rsrc.r != nil {
		if rc, rcok := rsrc.r.(io.Closer); rcok {
			err = rc.Close()
			rc = nil
		}
		rsrc.r = nil
	}
	if rsrc.reqst.firstrdr == rsrc {
		rsrc.reqst.firstrdr = rsrc.nextrsrc
		if rsrc.nextrsrc != nil {
			rsrc.nextrsrc.prvrsrc = nil
		}
	}
	if rsrc == rsrc.reqst.lastrdr {
		rsrc.reqst.lastrdr = rsrc.prvrsrc
	}

	if rsrc.prvrsrc != nil {
		rsrc.prvrsrc.nextrsrc = rsrc.nextrsrc
		rsrc.prvrsrc = nil
	}
	if rsrc.nextrsrc != nil {
		rsrc.nextrsrc = nil
	}
	if rsrc.reqst != nil {
		rsrc.reqst = nil
	}
	if rsrc.readBuffer != nil {
		rsrc.readBuffer = nil
	}
	if rsrc.rbuf != nil {
		rsrc.rbuf = nil
	}
	return
}

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

var callableResources map[string]func() io.Reader

func RegisterCallableResource(resource string, callable func() io.Reader, a ...interface{}) {

}
