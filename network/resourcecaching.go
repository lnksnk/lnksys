package network

import (
	active "github.com/efjoubert/lnksys/iorw/active"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ResourceInfo struct {
	*active.Active
	finfo    os.FileInfo
	path     string
	pathroot string
	modified time.Time
	rsSize   int64
}

func NextResourceInfo(rsrc *Resource, path string, pathroot string, finfo os.FileInfo) (nxtrsinfo *ResourceInfo) {
	nxtrsinfo = &ResourceInfo{path: path, pathroot: pathroot, finfo: finfo}
	if finfo != nil {
		nxtrsinfo.modified = finfo.ModTime()
		nxtrsinfo.rsSize = finfo.Size()
	}
	return
}

func (rsinfo *ResourceInfo) Path() string {
	return rsinfo.path
}

func (rsinfo *ResourceInfo) PathRoot() string {
	return rsinfo.pathroot
}

func (rsinfo *ResourceInfo) Close() {
	if rsinfo.Active != nil {
		rsinfo.Active.Close()
		rsinfo.Active = nil
	}
	if rsinfo.finfo != nil {
		rsinfo = nil
	}
}

func (rsinfo *ResourceInfo) IsActiveContent() (active bool) {
	var ext = filepath.Ext(rsinfo.path)
	if atvExtns != nil {
		active, _ = atvExtns[ext]
	}
	return
}

func (rsinfo *ResourceInfo) Reader(rsrc *Resource) (r io.Reader) {
	if rsinfo.finfo != nil {
		if strings.HasSuffix(rsinfo.pathroot, "/") && rsinfo.pathroot != "/" {
			r, _ = os.Open(rsinfo.pathroot[:len(rsinfo.pathroot)-1] + rsinfo.path)
		} else {
			r, _ = os.Open(rsinfo.pathroot + rsinfo.path)
		}
	}
	return
}

type ResourceInfoHandler struct {
}

func (rsifihndlr *ResourceInfoHandler) nextResourceRoots(reqst *Request, resourcepath string) (nxtrspath string, rmningrspath string) {
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

type rootType int

type RootHandler struct {
	rootpath  string
	resources map[string]*ResourceInfoHandler
}

func newRootHandler(root string) (rthndlr *RootHandler) {
	if root != "" {
		if fi, fierr := os.Stat(root); fierr == nil {
			if fi.IsDir() {
				if !strings.HasSuffix(root, "/") {
					root += "/"
				}
			} else {

			}
		}
	}
	return
}

type ResourcesManager struct {
	roots map[string]*RootHandler
}

func newResourcesManager() (rsrcsmngr *ResourcesManager) {
	rsrcsmngr = &ResourcesManager{roots: map[string]*RootHandler{}}
	return
}

func (rsrcsmngr *ResourcesManager) MapResourceRoot(a ...interface{}) {
	for len(a) >= 2 && len(a)%2 == 0 {
		if rspath, rspathok := a[0].(string); rspathok {
			if rspath != "" {
				if rsroot, rsrootok := a[1].(string); rsrootok {
					if rsroot != "" {

						a = a[2:]
					}
				} else {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}
}

var rsrcsmngr *ResourcesManager

func init() {
	if rsrcsmngr == nil {
		rsrcsmngr = newResourcesManager()
	}
}
