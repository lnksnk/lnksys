package network

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ResourceInfo struct {
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
