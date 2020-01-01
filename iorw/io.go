package io

import (
	"fmt"
	"io"
)

type Printing interface {
	Print(a ...interface{})
	Println(a ...interface{})
}

type ReaderWriter interface {
	io.Reader
	io.Seeker
	io.Closer
	io.Writer
}

type Reader struct {
	altR  io.Reader
	altRS io.ReadSeeker
	altRC io.ReadCloser
}

func NewReader(altR io.Reader) (r *Reader) {
	r = &Reader{}
	r.altR = altR
	if altRS, altRSOk := altR.(io.ReadSeeker); altRSOk {
		r.altRS = altRS
	}
	if altRC, altRCOk := altR.(io.ReadCloser); altRCOk {
		r.altRC = altRC
	}
	return
}

func (rd *Reader) Read(p []byte) (n int, err error) {
	if rd.altR != nil {
		n, err = rd.altR.Read(p)
	} else {
		err = io.EOF
	}
	return
}

func (rd *Reader) Seek(offset int64, whence int) (n int64, err error) {
	if rd.altRS != nil {
		n, err = rd.altRS.Seek(offset, whence)
	}
	return
}

func (r *Reader) Close() (err error) {
	if r.altRC != nil {
		err = r.altRC.Close()
	}
	return
}

type Writer struct {
	altW  io.Writer
	altWS io.WriteSeeker
	altWC io.WriteCloser
}

func NewWriter(altW io.Writer) (w *Writer) {
	w = &Writer{}
	w.altW = altW
	if altWS, altWSOk := altW.(io.WriteSeeker); altWSOk {
		w.altWS = altWS
	}
	if altWC, altWCOk := altW.(io.WriteCloser); altWCOk {
		w.altWC = altWC
	}
	return
}

func (wt *Writer) Write(p []byte) (n int, err error) {
	return
}

func (wt *Writer) Seek(offset int64, whence int) (n int64, err error) {
	if wt.altWS != nil {
		n, err = wt.altWS.Seek(offset, whence)
	}
	return
}

func (wt *Writer) Close() (err error) {
	if wt.altWC != nil {
		err = wt.altWC.Close()
	}
	return
}

type RW struct {
	*Reader
	*Writer
}

func NewRW(a interface{}) (rw *RW) {
	rw = &RW{}
	if altR, altROk := a.(io.Reader); altROk {
		rw.Reader = NewReader(altR)
	}
	if altW, altWOk := a.(io.Writer); altWOk {
		rw.Writer = NewWriter(altW)
	}
	return
}

func (rw *RW) Seek(offset int64, whence int) (n int64, err error) {
	if rw.Reader != nil {
		n, err = rw.Reader.Seek(offset, whence)
	}
	return
}

func (rw *RW) Read(p []byte) (n int, err error) {
	if rw.Reader != nil {
		n, err = rw.Reader.Read(p)
	} else {
		err = io.EOF
	}
	return
}

func (rw *RW) Write(p []byte) (n int, err error) {
	if rw.Writer != nil {
		n, err = rw.Writer.Write(p)
	} else {
		err = io.EOF
	}
	return
}

func (rw *RW) Close() (err error) {
	if rw.Reader != nil {
		err = rw.Reader.Close()
	}
	if rw.Writer != nil {
		err = rw.Writer.Close()
	}
	return
}

type BufferedRW struct {
	altRW         ReaderWriter
	buffer        [][]byte
	bytes         []byte
	bytesi        int
	bytesl        int
	wbytes        []byte
	wbytesi       int
	bufferSize    int64
	maxBufferSize int64
	offset        int64
	bufRWActn     bufRWAction
	bufRWActnDone chan bool
}

func NewBufferedRW(maxBufferSize int64, altRW ReaderWriter) (bufRW *BufferedRW) {
	bufRW = &BufferedRW{altRW: altRW, maxBufferSize: maxBufferSize, bufRWActn: bufRWNoAction, bufRWActnDone: make(chan bool, 1)}
	return
}

func (bufRW *BufferedRW) Size() (n int64) {
	if bufRW.altRW != nil {

	} else {
		/*if len(bufRW.buffer) > 0 {
			for _, bf := range bufRW.buffer {
				n += int64(len(bf))
			}
		}
		if bufRW.wbytesi > 0 {
			n += int64(bufRW.wbytesi)
		}*/
		n = bufRW.bufferSize
	}
	return n
}

func queueNextBufRWAction(bufRWActn bufRWAction, bufRW *BufferedRW) {
	bufRW.bufRWActn = bufRWActn
	//bufRWQueue <- bufRW
	//<-bufRW.bufRWActnDone
	if bufRW.bufRWActn == bufRWCopyRead {
		io.CopyN(bufRW, bufRW.altRW, bufRW.maxBufferSize)
		bufRW.bufRWActn = bufRWNoAction
	}
}

func (bufRW *BufferedRW) String() (s string) {
	s = ""
	if bufRW.altRW == nil {
		if bufRW.bufferSize > 0 {
			if len(bufRW.buffer) > 0 {
				for _, bf := range bufRW.buffer {
					s += string(bf)
				}
			}
		}
		if bufRW.wbytesi > 0 {
			s += string(bufRW.wbytes[0:bufRW.wbytesi])
		}
	}
	return
}

func (bufRW *BufferedRW) Reset() {

}

func (bufRW *BufferedRW) Read(p []byte) (n int, err error) {
	var pl = len(p)
	for n < pl {
		if bufRW.bytesl == 0 || bufRW.bytesl > 0 && bufRW.bytesi == bufRW.bytesl {
			if bufRW.bytesi > 0 {
				bufRW.bytesi = 0
			}

			if bufRW.bufferSize == 0 {
				if bufRW.altRW != nil && bufRW.maxBufferSize > 0 {

					queueNextBufRWAction(bufRWCopyRead, bufRW)
				}
				if bufRW.wbytesi > 0 {
					if bufRW.buffer == nil {
						bufRW.buffer = [][]byte{}
					}
					bufRW.buffer = append(bufRW.buffer, make([]byte, bufRW.wbytesi))
					copy(bufRW.buffer[len(bufRW.buffer)-1], bufRW.wbytes[0:bufRW.wbytesi])
					bufRW.bufferSize += int64(bufRW.wbytesi)
					bufRW.wbytesi = 0
				}
			}

			if len(bufRW.buffer) > 0 {
				bufRW.bytes = bufRW.buffer[0][:]
				bufRW.bytesl = len(bufRW.bytes)
				bufRW.buffer[0] = nil
				if len(bufRW.buffer) > 1 {
					bufRW.buffer = bufRW.buffer[1:]
				} else {
					bufRW.buffer = nil
				}
			} else {
				bufRW.bytesl = 0
				break
			}
		}
		for n < pl && bufRW.bytesi < bufRW.bytesl {
			if (pl - n) >= (bufRW.bytesl - bufRW.bytesi) {
				var cpl = copy(p[n:n+(bufRW.bytesl-bufRW.bytesi)], bufRW.bytes[bufRW.bytesi:bufRW.bytesi+(bufRW.bytesl-bufRW.bytesi)])
				n += cpl
				bufRW.bytesi += cpl
				if bufRW.bufferSize >= int64(cpl) {
					bufRW.bufferSize -= int64(cpl)
				}
			} else if (pl - n) < (bufRW.bytesl - bufRW.bytesi) {
				var cpl = copy(p[n:n+(pl-n)], bufRW.bytes[bufRW.bytesi:bufRW.bytesi+(pl-n)])
				n += cpl
				bufRW.bytesi += cpl
				if bufRW.bufferSize >= int64(cpl) {
					bufRW.bufferSize -= int64(cpl)
				}
			}
		}
	}
	if n == 0 && err == nil {
		err = io.EOF
	}
	return
}

func (bufRW *BufferedRW) Println(a ...interface{}) {
	bufRW.Print(a...)
	bufRW.Print("\r\n")
}

func (bufRW *BufferedRW) Print(a ...interface{}) {
	for _, d := range a {
		if r, rok := d.(io.Reader); rok {
			io.Copy(bufRW, r)
		} else if uarr, uarrok := d.([]uint8); uarrok {
			fmt.Fprint(bufRW, string(uarr))
		} else if runearr, runearrok := d.([]rune); runearrok {
			fmt.Fprint(bufRW, string(runearr))
		} else if barr, barrok := d.([]byte); barrok {
			fmt.Fprint(bufRW, string(barr))
		} else {
			fmt.Fprint(bufRW, d)
		}
	}
}

func (bufRW *BufferedRW) Write(p []byte) (n int, err error) {
	var pl = len(p)
	for n < pl {
		if len(bufRW.wbytes) == bufRW.wbytesi {
			if bufRW.wbytesi > 0 {
				if bufRW.buffer == nil {
					bufRW.buffer = [][]byte{}
				}
				bufRW.buffer = append(bufRW.buffer, make([]byte, bufRW.wbytesi))
				copy(bufRW.buffer[len(bufRW.buffer)-1], bufRW.wbytes[0:bufRW.wbytesi])
				bufRW.bufferSize += int64(bufRW.wbytesi)
				bufRW.wbytesi = 0
			}
			if len(bufRW.wbytes) == 0 {
				bufRW.wbytes = make([]byte, 81920)
			}
		}
		for n < pl && bufRW.wbytesi < len(bufRW.wbytes) {
			if (pl - n) >= (len(bufRW.wbytes) - bufRW.wbytesi) {
				var cpl = copy(bufRW.wbytes[bufRW.wbytesi:bufRW.wbytesi+(len(bufRW.wbytes)-bufRW.wbytesi)], p[n:+(len(bufRW.wbytes)-bufRW.wbytesi)])
				n += cpl
				bufRW.wbytesi += cpl

			} else if (pl - n) < (len(bufRW.wbytes) - bufRW.wbytesi) {
				var cpl = copy(bufRW.wbytes[bufRW.wbytesi:bufRW.wbytesi+(pl-n)], p[n:n+(pl-n)])
				n += cpl
				bufRW.wbytesi += cpl
			}
		}
	}
	return
}

func (bufRW *BufferedRW) Close() (err error) {
	if bufRW.buffer != nil {
		for len(bufRW.buffer) > 0 {
			bufRW.buffer[0] = nil
			bufRW.buffer = bufRW.buffer[1:]
		}
		bufRW.buffer = nil
	}
	if bufRW.bytes != nil {
		bufRW.bytes = nil
	}
	if bufRW.wbytes != nil {
		bufRW.wbytes = nil
	}
	if bufRW.altRW != nil {
		bufRW.altRW = nil
	}
	if bufRW.bufRWActnDone != nil {
		close(bufRW.bufRWActnDone)
		bufRW.bufRWActnDone = nil
	}
	return
}

func (bufRW *BufferedRW) Seek(offset int64, whence int) (n int64, err error) {
	if bufRW.altRW != nil {
		n, err = bufRW.altRW.Seek(offset, whence)
	} else {
		if bufRW.bytesi > 0 && bufRW.bytesi < bufRW.bytesl {
			if len(bufRW.wbytes) == 0 {
				bufRW.wbytes = make([]byte, 81920)
			}
			copy(bufRW.wbytes[0:(bufRW.bytesl-bufRW.bytesi)], bufRW.bytes[bufRW.bytesi:bufRW.bytesi+(bufRW.bytesl-bufRW.bytesi)])
			bufRW.wbytesi = (bufRW.bytesl - bufRW.bytesi)
		}
		if bufRW.wbytesi > 0 {
			if bufRW.buffer == nil {
				bufRW.buffer = [][]byte{}
			}
			bufRW.buffer = append(bufRW.buffer, bufRW.wbytes[0:bufRW.wbytesi])
			bufRW.bufferSize += int64(bufRW.wbytesi)
			bufRW.wbytesi = 0
		}
	}
	return
}

type bufRWAction int

const (
	bufRWNoAction bufRWAction = 0
	bufRWCopyRead bufRWAction = 0
)

var bufRWQueue chan *BufferedRW

func init() {
	if bufRWQueue == nil {
		bufRWQueue = make(chan *BufferedRW)
		go func() {
			for {
				select {
				case bufRW := <-bufRWQueue:
					go func() {
						if bufRW.bufRWActn == bufRWCopyRead {
							io.CopyN(bufRW, bufRW.altRW, bufRW.maxBufferSize)
							bufRW.bufRWActn = bufRWNoAction
						}
						bufRW.bufRWActnDone <- true
					}()
				}
			}
		}()
	}
}