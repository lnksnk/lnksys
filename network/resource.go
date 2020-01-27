package network

import (
	"archive/zip"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	embed "github.com/efjoubert/lnksys/embed"
	iorw "github.com/efjoubert/lnksys/iorw"
)

type Resource struct {
	reqst           *Request
	finfo           os.FileInfo
	r               io.Reader
	path            string
	pathroot        string
	size            int64
	readBuffer      []byte
	readBufferi     int
	readBufferl     int
	readRuneBuffer  []rsrRune
	readRuneBufferi int
	readRuneBufferl int
	rbuf            *bufio.Reader
	activeInverse   bool
	activeEnd       bool
	isfirst         bool
	disableActive   bool
	pipeR           *io.PipeReader
	pipeW           *io.PipeWriter
}

type rsrRune struct {
	rsrrerr error
	rsrsize int
	rsrr    rune
}

func (rsrr rsrRune) ReadRune() (rune, int, error) {
	return rsrr.rsrr, rsrr.rsrsize, rsrr.rsrrerr
}

func newRsrRune(r rune, size int, err error) rsrRune {
	return rsrRune{rsrrerr: err, rsrsize: size, rsrr: r}
}

func (rsrc *Resource) ReadRune() (r rune, size int, err error) {
	if rsrc.rbuf == nil {
		rsrc.pipeR, rsrc.pipeW = io.Pipe()
		if rsrc.r == nil && rsrc.finfo != nil {
			if strings.HasSuffix(rsrc.pathroot, "/") && rsrc.pathroot != "/" {
				rsrc.r, _ = os.Open(rsrc.pathroot[:len(rsrc.pathroot)-1] + rsrc.path)
			} else {
				rsrc.r, _ = os.Open(rsrc.pathroot + rsrc.path)
			}
		}
		go func(rr io.Reader) {
			defer rsrc.pipeW.Close()
			for {
				ws, wserr := io.Copy(rsrc.pipeW, rsrc.r)
				if ws == 0 || wserr != nil {
					break
				}
			}
		}(rsrc.r)
		rsrc.rbuf = bufio.NewReader(rsrc.pipeR)
	}
	if rsrc.readRuneBufferi == 0 || (rsrc.readRuneBufferl > 0 && rsrc.readRuneBufferi == rsrc.readRuneBufferl) {
		if rsrc.readRuneBufferi > 0 {
			rsrc.readRuneBufferi = 0
		}
		if rsrc.readRuneBuffer == nil {
			rsrc.readRuneBuffer = make([]rsrRune, 81920)
		}
		rsrc.readRuneBufferl = 0
		for {
			rr, rsize, rerr := rsrc.rbuf.ReadRune()
			if rsize > 0 {
				rsrc.readRuneBuffer[rsrc.readRuneBufferl] = newRsrRune(rr, rsize, rerr)
				rsrc.readRuneBufferl++
			}
			if rerr != nil || rsrc.readRuneBufferl == len(rsrc.readRuneBuffer) {
				r = 0
				size = 0
				err = nil
				break
			}
		}
		if rsrc.readRuneBufferl == 0 {
			r = 0
			size = 0
			err = io.EOF
			return
		}
	}
	if rsrc.readRuneBufferi < rsrc.readRuneBufferl {
		r, size, err = rsrc.readRuneBuffer[rsrc.readRuneBufferi].ReadRune()
		rsrc.readRuneBufferi++
	}
	//r, size, err = rsrc.rbuf.ReadRune()
	return
}

func (rsrc *Resource) IsActiveContent() (active bool) {
	var ext = filepath.Ext(rsrc.path)
	if atvExtns != nil {
		active, _ = atvExtns[ext]
	}
	return
}

func (reqst *Request) nextResourceRoots(resourcepath string) (nxtrspath string, rmningrspath string) {
	if len(reqst.rootpaths) > 0 && resourcepath != "" {
		nxtrspath = ""
		rmningrspath = ""
		for _, respath := range reqst.rootpaths {
			if respath != "" {
				if _, rspathok := roots[respath]; rspathok {
					if strings.HasPrefix(resourcepath, "/") {
						if strings.HasPrefix(respath, "/") && strings.HasPrefix(resourcepath, respath) {
							if len(respath) > len(nxtrspath) {
								nxtrspath = respath
								rmningrspath = resourcepath[len(respath):]
							}
						} else if strings.HasPrefix(resourcepath, "/"+respath) {
							if len(respath) > len(nxtrspath) {
								nxtrspath = respath
								rmningrspath = resourcepath[len(respath)+1:]
							}
						}
					} else if strings.HasPrefix(resourcepath, respath) {
						if len(respath) > len(nxtrspath) {
							nxtrspath = respath
							rmningrspath = resourcepath[len(respath):]
						}
					}
				}
			}
		}
	}
	return
}

func NewResource(reqst *Request, resourcepath string) (rsrc *Resource) {

	var r io.Reader = nil

	var nxtrspath, rmningrspath = reqst.nextResourceRoots(resourcepath)

	var finfo os.FileInfo = nil
	var lastPathRoot = ""
	var disableActive = false

	var findR = func(rspathrt string, rspath string) (rf io.Reader) {
		if rf = embed.EmbedFindJS(rspath); rf == nil {
			var rootFound = roots[rspathrt]
			var pathDelim = "/"
			if strings.HasPrefix(rootFound, "http:") || strings.HasPrefix(rootFound, "https:") {
				var qryparams = ""
				if disableActive {
					qryparams = "disable-active=Y"
				}
				if strings.LastIndex(rspath, "?") > -1 {
					qryparams = rspath[strings.LastIndex(rspath, "?")+1:]
					rspath = rspath[:strings.LastIndex(rspath, "?")]
				}
				if strings.LastIndex(rootFound, "?") > -1 {
					if qryparams == "" {
						qryparams = rootFound[strings.LastIndex(rootFound, "?")+1:]
					} else {
						qryparams = qryparams + "&" + rootFound[strings.LastIndex(rootFound, "?")+1:]
					}
					rootFound = rootFound[:strings.LastIndex(rootFound, "?")]
				}
				if strings.HasPrefix(rspath, "/") {
					pathDelim = ""
				}
				if strings.HasSuffix(rootFound, "/") {
					rootFound = rootFound[:len(rootFound)-1]
				}
				var tlkr = NewTalker()
				var rw = iorw.NewBufferedRW(81920)
				if qryparams != "" {
					qryparams = "?" + qryparams
				}
				var tlkrhdrs = map[string][]string{}
				var tlkrparams = map[string][]string{}
				if reqst.isfirstResource {
					tlkr.FSend(rw, reqst.RequestContent(), tlkrhdrs, rootFound+pathDelim+rspath+qryparams, tlkrparams, reqst.params)
				} else {
					tlkr.FSend(rw, nil, tlkrhdrs, rootFound+pathDelim+rspath+qryparams, tlkrparams)
				}
				tlkr.Close()
				rf = rw
			} else if rffi, rffierr := os.Stat(rootFound); rffierr == nil && rffi.IsDir() {
				if strings.HasSuffix(rootFound, "/") && pathDelim == "/" {
					pathDelim = ""
				}
				if !func() bool {
					var resource = rootFound + ""
					var tmprestest = rspath
					if strings.HasPrefix(tmprestest, "/") {
						tmprestest = tmprestest[1:]
					}
					if strings.HasSuffix(tmprestest, "/") {
						tmprestest = tmprestest[:len(tmprestest)-1]
					}

					resource = resource + pathDelim + tmprestest
					if fi, fierr := os.Stat(resource); fierr == nil {
						if !fi.IsDir() {
							lastPathRoot = rootFound
							finfo = fi
							return true
						}
					}
					return false
				}() {
					var ressplit = strings.Split(rspath, "/")
					var tmpres = ""
					for nrs := range ressplit {
						if nrs > 0 {
							tmpres = strings.Join(ressplit[:nrs], "/")
							var zipresource = rootFound + pathDelim + tmpres[:len(tmpres)] + ".zip"
							if _, fiziperr := os.Stat(zipresource); fiziperr == nil {
								func() {
									if zipr, ziprerr := zip.OpenReader(zipresource); ziprerr == nil {
										for _, f := range zipr.File {
											if f.Name == strings.Join(ressplit[nrs:], "/") {
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
							}
							if rf != nil || finfo != nil {
								break
							}
						}
					}
				}
			} else if strings.HasSuffix(rootFound, ".zip") {
				if _, fiziperr := os.Stat(rootFound); fiziperr == nil {
					if zipr, ziprerr := zip.OpenReader(rootFound); ziprerr == nil {
						for _, f := range zipr.File {
							if f.Name == rspath {
								if ziprrc, ziprrcerr := f.Open(); ziprrcerr == nil {
									rf = ziprrc
									finfo = f.FileInfo()
									break
								}
								break
							}
						}
					}
					if rf != nil || finfo != nil {
						return
					}
				}
			}
		}
		return
	}
	var activeInverse = false

	if rmningrspath != "" && nxtrspath != "" {
		if r = findR(nxtrspath, rmningrspath); r == nil && finfo == nil && strings.Count(rmningrspath, "@") > 0 && strings.Index(rmningrspath, "@") >= 0 && strings.Index(rmningrspath, "@") != strings.LastIndex(rmningrspath, "@") {
			activeInverse = true
			rmningrspath = strings.Replace(rmningrspath, "@", "", -1)
			if r = findR(nxtrspath, rmningrspath); r != nil || finfo != nil {
				resourcepath = rmningrspath
				if !strings.HasPrefix(resourcepath, "/") {
					resourcepath = "/" + resourcepath
				}
			}
		} else if r == nil && finfo == nil && strings.Count(rmningrspath, "!") > 0 && strings.Index(rmningrspath, "!") >= 0 && strings.Index(rmningrspath, "!") != strings.LastIndex(rmningrspath, "!") {
			disableActive = true
			rmningrspath = strings.Replace(rmningrspath, "!", "", -1)
			if r = findR(nxtrspath, rmningrspath); r != nil || finfo != nil {
				resourcepath = rmningrspath
				if !strings.HasPrefix(resourcepath, "/") {
					resourcepath = "/" + resourcepath
				}
			}
		}
		if r != nil || finfo != nil {
			resourcepath = rmningrspath
			if !strings.HasPrefix(resourcepath, "/") {
				resourcepath = "/" + resourcepath
			}
		}
	}
	if r != nil || finfo != nil {
		rsrc = &Resource{
			path:          resourcepath,
			pathroot:      lastPathRoot,
			r:             r,
			finfo:         finfo,
			reqst:         reqst,
			activeInverse: activeInverse,
			activeEnd:     false,
			isfirst:       reqst.isfirstResource,
			disableActive: disableActive}
		if reqst.isfirstResource {
			reqst.isfirstResource = false
		}
		if finfo != nil {
			rsrc.size = finfo.Size()
		}
		return
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

func (rsrc *Resource) internalRead(p []byte) (n int, err error) {
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

func (rsrc *Resource) Read(p []byte) (n int, err error) {
	n, err = rsrc.internalRead(p)
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
	if rsrc.reqst != nil {
		rsrc.reqst = nil
	}
	if rsrc.readBuffer != nil {
		rsrc.readBuffer = nil
	}
	if rsrc.readRuneBuffer != nil {
		rsrc.readRuneBuffer = nil
	}
	if rsrc.rbuf != nil {
		rsrc.rbuf = nil
	}
	return
}
