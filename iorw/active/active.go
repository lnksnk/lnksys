package active

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	goja "github.com/dop251/goja"
	gojaparse "github.com/dop251/goja/parser"
	iorw "github.com/efjoubert/lnksys/iorw"
)

type activeExecutor struct {
	passiveBuffer           [][]rune
	activeBuffer            [][]rune
	activeBufferOffset      int64
	lastActiveBufferOffset  int64
	hasCode                 bool
	foundCode               bool
	passiveBufferOffset     int64
	lastPassiveBufferOffset int64
	atvprsr                 *activeParser
	prgrm                   chan *goja.Program
	prgrmerr                chan error
	prgrmbufin              *bufio.Writer
	pipeprgrminw            *io.PipeWriter
	pipeprgrminr            *io.PipeReader
	pipeprgrmoutw           *io.PipeWriter
	pipeprgrmoutr           *io.PipeReader
}

func newActiveExecutor(atvprsr *activeParser) (atvxctr *activeExecutor) {
	atvxctr = &activeExecutor{atvprsr: atvprsr, foundCode: false, hasCode: false, passiveBufferOffset: 0, lastPassiveBufferOffset: 0, activeBufferOffset: 0, lastActiveBufferOffset: 0}
	return
}

func (atvxctr *activeExecutor) passiveBuf() [][]rune {
	if atvxctr.passiveBuffer == nil {
		atvxctr.passiveBuffer = [][]rune{}
	}
	return atvxctr.passiveBuffer
}

func (atvxctr *activeExecutor) captureActiveRunes(atvrnes []rune) {
	if len(atvrnes) > 0 {
		if atvxctr.pipeprgrminw == nil && atvxctr.pipeprgrminr == nil && atvxctr.prgrm == nil && atvxctr.prgrmerr == nil {
			atvxctr.prgrm = make(chan *goja.Program, 1)
			atvxctr.prgrmerr = make(chan error, 1)
			atvxctr.pipeprgrminr, atvxctr.pipeprgrminw = io.Pipe()
			atvxctr.pipeprgrmoutr, atvxctr.pipeprgrmoutw = io.Pipe()
			go func(pin *io.PipeReader, po *io.PipeWriter) {
				defer po.Close()
				go func() {
					defer atvxctr.pipeprgrmoutw.Close()
					var bytes = make([]byte, 8192)
					/*bfr := bufio.NewReader(atvxctr.pipeprgrminr)
					bfw := bufio.NewWriter(atvxctr.pipeprgrmoutw)
					code := ""
					func() {
						defer func() {
							if code != "" {
								fmt.Println(code)
							}
						}()
						for {
							r, s, e := bfr.ReadRune()
							if e != nil {
								if e == io.EOF {
									bfw.Flush()
								}
								break
							} else {
								if s > 0 {
									code += string(r)
									s, e = bfw.WriteRune(r)
									if e != nil {
										break
									}
								}
							}
						}
					}()*/
					io.CopyBuffer(atvxctr.pipeprgrmoutw, atvxctr.pipeprgrminr, bytes)
					bytes = nil
				}()
				var parsedprgm, parsedprgmerr = gojaparse.ParseFile(nil, "", atvxctr.pipeprgrmoutr, 0)

				if parsedprgmerr == nil {
					nxtprm, nxtprmerr := goja.CompileAST(parsedprgm, false)
					atvxctr.prgrm <- nxtprm
					atvxctr.prgrmerr <- nxtprmerr
				} else {
					atvxctr.prgrm <- nil
					atvxctr.prgrmerr <- parsedprgmerr
				}
			}(atvxctr.pipeprgrminr, atvxctr.pipeprgrminw)
		}
	}
	atvxctr.pipeprgrminw.Write([]byte(string(atvrnes)))
}

func (atvxctr *activeExecutor) activeBuf() [][]rune {
	if atvxctr.activeBuffer == nil {
		atvxctr.activeBuffer = [][]rune{}
	}
	return atvxctr.activeBuffer
}

func (atvxctr *activeExecutor) close() {
	if atvxctr.passiveBuffer != nil {
		for len(atvxctr.passiveBuffer) > 0 {
			atvxctr.passiveBuffer[0] = nil
			atvxctr.passiveBuffer = atvxctr.passiveBuffer[1:]
		}
		atvxctr.passiveBuffer = nil
	}
	if atvxctr.activeBuffer != nil {
		for len(atvxctr.activeBuffer) > 0 {
			atvxctr.activeBuffer[0] = nil
			atvxctr.activeBuffer = atvxctr.activeBuffer[1:]
		}
		atvxctr.activeBuffer = nil
	}
	if atvxctr.atvprsr != nil {
		atvxctr.atvprsr = nil
	}
	if atvxctr.prgrmbufin != nil {
		atvxctr.prgrmbufin = nil
	}
	if atvxctr.pipeprgrminw != nil {
		atvxctr.pipeprgrminw = nil
	}
	if atvxctr.pipeprgrminr != nil {
		atvxctr.pipeprgrminr = nil
	}
	if atvxctr.pipeprgrmoutw != nil {
		atvxctr.pipeprgrmoutw = nil
	}
	if atvxctr.pipeprgrmoutr != nil {
		atvxctr.pipeprgrmoutr = nil
	}
	if atvxctr.prgrmerr != nil {
		close(atvxctr.prgrmerr)
		atvxctr.prgrmerr = nil
	}
	if atvxctr.prgrm != nil {
		close(atvxctr.prgrm)
		atvxctr.prgrm = nil
	}
}

func (atvxctr *activeExecutor) PassivePrint(atv *Active, fromOffset int64, toOffset int64) {
	if len(atvxctr.passiveBuffer) > 0 {
		if fromOffset >= 0 && toOffset <= atvxctr.passiveBufferOffset {
			var psi = int(0)
			var pei = int(0)
			var pfrom = int64(0)
			var pto = int64(0)
			var pl = int64(0)
			for _, psvb := range atvxctr.passiveBuffer {
				pl = int64(len(psvb))
				pto = pl + pfrom
				if fromOffset < pto {
					if fromOffset < pfrom {
						psi = int(pfrom - fromOffset)
					} else {
						psi = int(fromOffset - pfrom)
					}
					if pto <= toOffset {
						pei = int(pl - (pto - toOffset))
						if atv != nil {
							atv.Print(string(psvb[psi:pei]))
						}
						if pto == toOffset {
							break
						}
					} else if toOffset < pto {
						if pto-toOffset > 0 {
							pei = int(pl - (pto - toOffset))
							if atv != nil {
								atv.Print(string(psvb[psi:pei]))
							}
						}
						break
					}
				}
				pfrom += pto
			}
		}
	}
}

type activeParser struct {
	atv        *Active
	atvrdr     *iorw.BufferedRW
	rdrRune    io.RuneReader
	rdskr      io.Seeker
	maxBufSize int64
	lck        *sync.Mutex
	//
	runesToParse  []rune
	runesToParsei int
	runeLabel     [][]rune
	runeLabelI    []int
	runePrvR      []rune
	//
	passiveRune    []rune
	passiveRunei   int
	disablePsvRune bool
	//
	activeRune  []rune
	activeRunei int

	parsingLevel int

	//
	psvRunesToParse  []rune
	psvRunesToParsei int
	psvLabel         [][]rune
	psvLabelI        []int
	psvPrvR          []rune
	//
	foundCdeTxt bool
	cdeTxt      rune
	pvrCdeTxt   rune
	atvxctr     []*activeExecutor
	xctngxctrs  []*activeExecutor
	atvrchan    chan rune
	atvrprcrone chan bool
}

func (atvprsr *activeParser) atvxctor(prsnglvl int) (atvxctr *activeExecutor) {
	if atvprsr.atvxctr == nil {
		atvprsr.atvxctr = []*activeExecutor{}
	}
	if len(atvprsr.atvxctr) < prsnglvl+1 {
		atvprsr.atvxctr = append(atvprsr.atvxctr, newActiveExecutor(atvprsr))
	}
	atvxctr = atvprsr.atvxctr[prsnglvl]
	return
}

func (atvprsr *activeParser) atvbufrdr() *iorw.BufferedRW {
	if atvprsr.atvrdr == nil {
		atvprsr.atvrdr = iorw.NewBufferedRW(atvprsr.maxBufSize, nil)
	}
	return atvprsr.atvrdr
}

func (atvprsr *activeParser) Reset() {
	if len(atvprsr.runeLabel) > 0 {
		atvprsr.runeLabelI[0] = 0
		atvprsr.runeLabelI[1] = 0
	}
	if len(atvprsr.runePrvR) == 1 {
		atvprsr.runePrvR[0] = rune(0)
	}
	if atvprsr.runesToParsei > 0 {
		atvprsr.runesToParsei = 0
	}
}

func (atvprsr *activeParser) Close() {
	if len(atvprsr.runeLabel) > 0 {
		atvprsr.runeLabelI = nil
		atvprsr.runeLabel = nil
	}
	if len(atvprsr.runePrvR) == 1 {
		atvprsr.runePrvR = nil
	}
	if atvprsr.runesToParsei > 0 {
		atvprsr.runesToParsei = 0
	}
	if len(atvprsr.runesToParse) > 0 {
		atvprsr.runesToParse = nil
	}
	if atvprsr.rdrRune != nil {
		atvprsr.rdrRune = nil
	}
	if atvprsr.rdskr != nil {
		atvprsr.rdskr = nil
	}
	//
	if atvprsr.runesToParse != nil {
		atvprsr.runesToParse = nil
	}
	if atvprsr.runeLabel != nil {
		atvprsr.runeLabel = nil
	}
	if atvprsr.runeLabelI != nil {
		atvprsr.runeLabelI = nil
	}
	if atvprsr.runePrvR != nil {
		atvprsr.runePrvR = nil
	}
	//
	if atvprsr.passiveRune != nil {
		atvprsr.passiveRune = nil
	}
	//
	if atvprsr.activeRune != nil {
		atvprsr.activeRune = nil
	}

	//
	if atvprsr.psvRunesToParse != nil {
		atvprsr.psvRunesToParse = nil
	}
	if atvprsr.psvLabel != nil {
		atvprsr.psvLabel = nil
	}
	if atvprsr.psvLabelI != nil {
		atvprsr.psvLabelI = nil
	}
	if atvprsr.psvPrvR != nil {
		atvprsr.psvPrvR = nil
	}
	//
	if atvprsr.atvxctr != nil {
		for len(atvprsr.atvxctr) > 0 {
			atvprsr.atvxctr[0].close()
			atvprsr.atvxctr[0] = nil
			atvprsr.atvxctr = atvprsr.atvxctr[1:]
		}
		atvprsr.atvxctr = nil
	}
	if atvprsr.atv != nil {
		atvprsr.atv = nil
	}
	if atvprsr.lck != nil {
		atvprsr.lck = nil
	}
	if atvprsr.atvrchan != nil {
		close(atvprsr.atvrchan)
	}
	if atvprsr.atvrprcrone != nil {
		close(atvprsr.atvrprcrone)
	}
}

func (atvprsr *activeParser) APrint(a ...interface{}) (err error) {
	if len(a) > 0 {
		atvprsr.lck.Lock()
		defer atvprsr.lck.Unlock()
		var canCheckDone = false
		var prcrune = func(rn rune) {
			processRune(atvprsr.parsingLevel, rn, atvprsr, atvprsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
		}
		var stopReading = false
		for _, d := range a {
			if rnrd, rnrdrok := d.(io.RuneReader); rnrdrok {
				if atvprsr.atvrdr != nil {
					for {
						if rne, rnsize, rnerr := atvprsr.atvrdr.ReadRune(); rnerr == nil {
							if rnsize > 0 {
								prcrune(rne)
							}
						} else {
							if rnerr != io.EOF {
								err = rnerr
								stopReading = true
							}
							break
						}
					}
				}
				if stopReading {
					break
				}
				for {
					if rne, rnsize, rnerr := rnrd.ReadRune(); rnerr == nil {
						if rnsize > 0 {
							prcrune(rne)
						}
					} else {
						if rnerr != io.EOF {
							err = rnerr
							stopReading = true
						}
						break
					}
				}
			} else {
				atvprsr.atvbufrdr().Print(d)
			}
			if stopReading {
				break
			}
		}
		if atvprsr.atvrdr != nil {
			for {
				if rne, rnsize, rnerr := atvprsr.atvrdr.ReadRune(); rnerr == nil {
					if rnsize > 0 {
						prcrune(rne)
					}
				} else {
					if rnerr != io.EOF {
						err = rnerr
					}
					break
				}
			}
		}
		if canCheckDone {
			atvprsr.atvrprcrone <- true
			<-atvprsr.atvrprcrone
		}
	}
	return
}

func cPrint(a ...interface{}) {
	if len(a) > 0 {
		var cbuf = iorw.NewBufferedRW(8192, nil)
		cbuf.Print(a...)
		fmt.Print(cbuf.String())
		cbuf.Close()
		cbuf = nil
	}
}

func preppingActiveParsing(atvprsr *activeParser) (atvxctr *activeExecutor) {
	flushPassiveContent(func() *activeExecutor { return atvprsr.atvxctor(atvprsr.parsingLevel) }, atvprsr, true)
	if len(atvprsr.atvxctr) > atvprsr.parsingLevel {
		if atvprsr.atvxctr[atvprsr.parsingLevel].foundCode {
			atvxctr = atvprsr.atvxctr[atvprsr.parsingLevel]
			flushActiveCode(func() *activeExecutor {
				return atvxctr
			}, atvprsr, true)
			if atvxctr.pipeprgrminw != nil {
				atvxctr.pipeprgrminw.Close()
			}
		} else {
			atvxctr = atvprsr.atvxctr[atvprsr.parsingLevel]
		}
	}

	atvprsr.parsingLevel++
	if atvprsr.runesToParsei > 0 {
		atvprsr.runesToParsei = 0
	}
	if atvprsr.runesToParse != nil {
		atvprsr.runesToParse = nil
	}
	if len(atvprsr.runeLabelI) == 2 {
		atvprsr.runeLabelI[0] = 0
		atvprsr.runeLabelI[1] = 0
	}
	if len(atvprsr.runePrvR) == 1 {
		atvprsr.runePrvR[0] = 0
	}
	if atvprsr.passiveRune != nil {
		atvprsr.passiveRune = nil
	}
	if atvprsr.activeRune != nil {
		atvprsr.passiveRune = nil
	}
	if atvprsr.activeRunei > 0 {
		atvprsr.activeRunei = 0
	}
	if len(atvprsr.psvLabelI) == 2 {
		atvprsr.psvLabelI[0] = 0
		atvprsr.psvLabelI[1] = 0
	}
	if len(atvprsr.psvPrvR) == 1 {
		atvprsr.psvPrvR[0] = 0
	}
	if atvprsr.foundCdeTxt {
		atvprsr.foundCdeTxt = false
	}
	atvprsr.pvrCdeTxt = rune(0)
	atvprsr.cdeTxt = rune(0)
	if atvxctr.foundCode {
		if atvprsr.xctngxctrs == nil {
			atvprsr.xctngxctrs = []*activeExecutor{}
		}
		atvprsr.xctngxctrs = append(atvprsr.xctngxctrs, atvxctr)
	}
	return atvxctr
}

func wrappingupActiveParsing(atvprsr *activeParser) {
	if atvprsr.parsingLevel > 0 {
		atvprsr.parsingLevel--

		if atvprsr.atvxctor(atvprsr.parsingLevel).foundCode {
			atvprsr.atvxctor(atvprsr.parsingLevel).foundCode = false
		}
		for len(atvprsr.atvxctr) > atvprsr.parsingLevel {
			var psvbufi = len(atvprsr.atvxctr) - 1
			atvprsr.atvxctr[psvbufi].close()
			atvprsr.atvxctr[psvbufi] = nil
			atvprsr.atvxctr = atvprsr.atvxctr[:psvbufi]
		}
	}
}

func (atvprsr *activeParser) ACommit(a ...interface{}) (acerr error) {
	if len(a) > 0 {
		acerr = atvprsr.APrint(a...)
	}
	if acerr == nil {
		defer func() {
			if err := recover(); err != nil {
				acerr = fmt.Errorf("panic: %+v\n ", err)
			}
			wrappingupActiveParsing(atvprsr)
			//atvprsr.lck.Unlock()
		}()
		if atvxctr := preppingActiveParsing(atvprsr); atvxctr != nil && atvxctr.foundCode {

			if atvprsr.atv != nil {
				if atvprsr.atv.vm == nil {
					atvprsr.atv.vm = goja.New()
					if atvprsr.atv.callinclude != nil {
						atvprsr.atv.vm.Set("include", atvprsr.atv.callinclude)
					}
					atvprsr.atv.vm.Set("out", atvprsr.atv)
					atvprsr.atv.vm.Set("CPrint", func(a ...interface{}) {
						cPrint(a...)
					})
					atvprsr.atv.vm.Set("CPrintln", func(a ...interface{}) {
						cPrint(a...)
						cPrint("\r\n")
					})
					atvprsr.atv.vm.Set("PassivePrint", func(fromOffset int64, toOffset int64) {
						atvprsr.PassivePrint(fromOffset, toOffset)
					})

					if len(activeGlobalMap) > 0 {
						for k, v := range activeGlobalMap {
							if atvprsr.atv.vm.Get(k) != v {
								atvprsr.atv.vm.Set(k, v)
							}
						}
					}
				}
				if len(atvprsr.atv.activeMap) > 0 {
					for k, v := range atvprsr.atv.activeMap {
						if atvprsr.atv.vm.Get(k) != v {
							atvprsr.atv.vm.Set(k, v)
						}
					}
				}

				var nxtprm *goja.Program = nil

				nxtprm = <-atvxctr.prgrm
				acerr = <-atvxctr.prgrmerr

				if acerr == nil && nxtprm != nil {
					var _, vmerr = atvprsr.atv.vm.RunProgram(nxtprm)
					if vmerr != nil {
						fmt.Println(vmerr)
						acerr = vmerr
						atvprsr.atv.vm.Interrupt(acerr)
					}
				}
				if len(atvprsr.xctngxctrs) > 0 {
					if atvprsr.xctngxctrs[len(atvprsr.xctngxctrs)-1] == atvxctr {
						if len(atvprsr.xctngxctrs) > 1 {
							atvprsr.xctngxctrs = atvprsr.xctngxctrs[:len(atvprsr.xctngxctrs)-1]
						} else {
							atvprsr.xctngxctrs = nil
						}
					}
				}
			}
		}
	}
	return
}

func (atvprsr *activeParser) PassivePrint(fromOffset int64, toOffset int64) {
	if len(atvprsr.xctngxctrs) > 0 {
		atvprsr.xctngxctrs[len(atvprsr.xctngxctrs)-1].PassivePrint(atvprsr.atv, fromOffset, toOffset)
	}
}

func (atvprsr *activeParser) Print(a ...interface{}) {
	if atvprsr.atv != nil {
		atvprsr.atv.Print(a...)
	}
}

type Active struct {
	printer     iorw.Printing
	atvprsr     *activeParser
	vm          *goja.Runtime
	activeMap   map[string]interface{}
	callinclude func(a ...interface{}) interface{}
}

func (atv *Active) APrint(a ...interface{}) (err error) {
	if atv.atvprsr != nil {
		err = atv.atvprsr.APrint(a...)
	}
	return
}

func (atv *Active) ACommit(a ...interface{}) (err error) {
	if atv.atvprsr != nil {
		err = atv.atvprsr.ACommit(a...)
	}
	return
}

func (atv *Active) APrintln(a ...interface{}) {
	atv.APrint(a...)
	atv.APrint("/r/n")
}

func capturePassiveContent(curatvxctr func() *activeExecutor, atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	if pl > 0 {
		atvxctr := curatvxctr()
		for n < pl {
			if atvxctr.foundCode {
				if len(atvprsr.passiveRune) == 0 {
					atvprsr.passiveRune = make([]rune, 81920)
				}
				if n < pl && atvprsr.passiveRunei < len(atvprsr.passiveRune) {
					if (pl - n) >= (len(atvprsr.passiveRune) - atvprsr.passiveRunei) {
						var cl = copy(atvprsr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)], p[n:n+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)])
						atvprsr.passiveRunei += cl
						n += cl
						atvxctr.passiveBufferOffset += int64(cl)
					} else if (pl - n) < (len(atvprsr.passiveRune) - atvprsr.passiveRunei) {
						var cl = copy(atvprsr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(pl-n)], p[n:n+(pl-n)])
						atvprsr.passiveRunei += cl
						n += cl
						atvxctr.passiveBufferOffset += int64(cl)
					}
					if len(atvprsr.passiveRune) == atvprsr.passiveRunei {
						var psvRunes = make([]rune, atvprsr.passiveRunei)
						copy(psvRunes, atvprsr.passiveRune[0:atvprsr.passiveRunei])
						atvxctr.passiveBuf()
						atvxctr.passiveBuffer = append(atvxctr.passiveBuffer, psvRunes)
						psvRunes = nil
						atvprsr.passiveRunei = 0
					}
				} else {
					break
				}
			} else {
				atvprsr.atv.Print(string(p))
				n += pl
			}
		}
	}
	return
}

func flushPassiveContent(curatvxctr func() *activeExecutor, atvprsr *activeParser, force bool) {
	if atvprsr.runesToParsei > 0 {
		processUnparsedPassiveContent(curatvxctr, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
		atvprsr.runesToParsei = 0
	}

	if atvprsr.psvRunesToParsei > 0 {
		capturePassiveContent(curatvxctr, atvprsr, atvprsr.psvRunesToParse[0:atvprsr.psvRunesToParsei])
		atvprsr.psvRunesToParsei = 0
	}
	atvxctr := curatvxctr()
	if atvxctr.foundCode {
		if force {
			if atvprsr.passiveRunei > 0 {
				var psvRunes = make([]rune, atvprsr.passiveRunei)
				copy(psvRunes, atvprsr.passiveRune[0:atvprsr.passiveRunei])
				atvxctr.passiveBuf()
				atvxctr.passiveBuffer = append(atvxctr.passiveBuffer, psvRunes)
				psvRunes = nil
				atvprsr.passiveRunei = 0
			}
		}

		if atvxctr.lastPassiveBufferOffset < atvxctr.passiveBufferOffset {
			for _, arune := range []rune(fmt.Sprintf("PassivePrint(%d,%d);", atvxctr.lastPassiveBufferOffset, atvxctr.passiveBufferOffset)) {
				if len(atvprsr.runesToParse) == 0 {
					atvprsr.runesToParse = make([]rune, 81920)
				}
				atvprsr.runesToParse[atvprsr.runesToParsei] = arune
				atvprsr.runesToParsei++
				if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
					captureActiveCode(func() *activeExecutor {
						return atvxctr
					}, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
			}
			atvxctr.lastPassiveBufferOffset = atvxctr.passiveBufferOffset
		}
	}
}

func processUnparsedPassiveContent(curatvxctr func() *activeExecutor, atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	if pl > 0 {
		flushActiveCode(curatvxctr, atvprsr, false)
	}
	if pl > 0 {
		for n < pl && atvprsr.psvRunesToParsei < len(atvprsr.psvRunesToParse) {
			if (pl - n) >= (len(atvprsr.psvRunesToParse) - atvprsr.psvRunesToParsei) {
				var cl = copy(atvprsr.psvRunesToParse[atvprsr.psvRunesToParsei:atvprsr.psvRunesToParsei+(len(atvprsr.psvRunesToParse)-atvprsr.psvRunesToParsei)], p[n:n+(len(atvprsr.psvRunesToParse)-atvprsr.psvRunesToParsei)])
				n += cl
				atvprsr.psvRunesToParsei += cl
			} else if (pl - n) < (len(atvprsr.psvRunesToParse) - atvprsr.psvRunesToParsei) {
				var cl = copy(atvprsr.psvRunesToParse[atvprsr.psvRunesToParsei:atvprsr.psvRunesToParsei+(pl-n)], p[n:n+(pl-n)])
				n += cl
				atvprsr.psvRunesToParsei += cl
			}
			if atvprsr.psvRunesToParsei > 0 && atvprsr.psvRunesToParsei == len(atvprsr.psvRunesToParse) {
				capturePassiveContent(curatvxctr, atvprsr, atvprsr.psvRunesToParse[0:atvprsr.psvRunesToParsei])
				atvprsr.psvRunesToParsei = 0
			}
		}
	}
	return
}

func processRune(processlvl int, rne rune, atvprsr *activeParser, runelbl [][]rune, runelbli []int, runePrvR []rune) {
	var atvxctr *activeExecutor = nil
	var curatvxctr = func() *activeExecutor {
		atvxctr = atvprsr.atvxctor(processlvl)
		return atvxctr
	}

	if (!atvprsr.foundCdeTxt) && (runelbli[1] == 0 && runelbli[0] < len(runelbl[0])) {
		if runelbli[0] > 0 && runelbl[0][runelbli[0]-1] == runePrvR[0] && runelbl[0][runelbli[0]] != rne {
			processUnparsedPassiveContent(curatvxctr, atvprsr, runelbl[0][0:runelbli[0]])
			runelbli[0] = 0
			runePrvR[0] = rune(0)
		}
		if runelbl[0][runelbli[0]] == rne {
			runelbli[0]++
			if len(runelbl[0]) == runelbli[0] {
				curatvxctr().hasCode = false
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[0] > 0 {
				processUnparsedPassiveContent(curatvxctr, atvprsr, runelbl[0][0:runelbli[0]])
				runelbli[0] = 0
			}
			runePrvR[0] = rne
			processUnparsedPassiveContent(curatvxctr, atvprsr, runePrvR)
		}
	} else if (atvprsr.foundCdeTxt) || (runelbli[0] == len(runelbl[0]) && runelbli[1] < len(runelbl[1])) {
		if runelbli[1] > 0 && runelbl[1][runelbli[1]-1] == runePrvR[0] && runelbl[1][runelbli[1]] != rne {
			processUnparsedActiveCode(curatvxctr, atvprsr, runelbl[1][0:runelbli[1]])
			runelbli[1] = 0
			runePrvR[0] = rune(0)
		}
		if (!atvprsr.foundCdeTxt) && runelbl[1][runelbli[1]] == rne {
			runelbli[1]++
			if runelbli[1] == len(runelbl[1]) {
				if atvprsr.runesToParsei > 0 {
					captureActiveCode(curatvxctr, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
				runePrvR[0] = rune(0)
				runelbli[0] = 0
				runelbli[1] = 0
				curatvxctr().hasCode = false
				curatvxctr().lastPassiveBufferOffset = curatvxctr().passiveBufferOffset
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[1] > 0 {
				processUnparsedActiveCode(curatvxctr, atvprsr, runelbl[1][0:runelbli[1]])
				runelbli[1] = 0
			}
			runePrvR[0] = rne
			processUnparsedActiveCode(curatvxctr, atvprsr, runePrvR)
		}
	}
}

func captureActiveCode(curatvxctr func() *activeExecutor, atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	if pl > 0 {
		for n < pl {
			if len(atvprsr.activeRune) == 0 {
				atvprsr.activeRune = make([]rune, 81920)
			}
			if n < pl && atvprsr.activeRunei < len(atvprsr.activeRune) {
				if (pl - n) >= (len(atvprsr.activeRune) - atvprsr.activeRunei) {
					var cl = copy(atvprsr.activeRune[atvprsr.activeRunei:atvprsr.activeRunei+(len(atvprsr.activeRune)-atvprsr.activeRunei)], p[n:n+(len(atvprsr.activeRune)-atvprsr.activeRunei)])
					atvprsr.activeRunei += cl
					n += cl
					curatvxctr().activeBufferOffset += int64(cl)
				} else if (pl - n) < (len(atvprsr.activeRune) - atvprsr.activeRunei) {
					var cl = copy(atvprsr.activeRune[atvprsr.activeRunei:atvprsr.activeRunei+(pl-n)], p[n:n+(pl-n)])
					atvprsr.activeRunei += cl
					n += cl
					curatvxctr().activeBufferOffset += int64(cl)
				}
				if len(atvprsr.activeRune) == atvprsr.activeRunei {
					curatvxctr().captureActiveRunes(atvprsr.activeRune[0:atvprsr.activeRunei])
					atvprsr.activeRunei = 0
				}
			} else {
				break
			}
		}
	}
	return
}

func flushActiveCode(curatvxctr func() *activeExecutor, atvprsr *activeParser, force bool) {
	if atvprsr.runesToParsei > 0 {
		captureActiveCode(curatvxctr, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
		atvprsr.runesToParsei = 0
	}
	if force {
		if atvprsr.activeRunei > 0 {
			curatvxctr().captureActiveRunes(atvprsr.activeRune[0:atvprsr.activeRunei])
			atvprsr.activeRunei = 0
		}
	}
}

func processUnparsedActiveCode(curatvxctr func() *activeExecutor, atvprsr *activeParser, p []rune) (err error) {
	if len(p) > 0 {
		atvxctr := curatvxctr()

		for _, arune := range p {
			if atvprsr.foundCdeTxt {
				if atvprsr.pvrCdeTxt != rune('\\') && (atvprsr.cdeTxt == rune('"') || atvprsr.cdeTxt == rune('\'')) && atvprsr.cdeTxt == arune {
					atvprsr.foundCdeTxt = false
					atvprsr.cdeTxt = rune(0)
				}
			} else {
				if atvprsr.pvrCdeTxt != rune('\\') && arune == rune('"') || arune == rune('\'') {
					atvprsr.cdeTxt = arune
					atvprsr.foundCdeTxt = true
				}
			}
			atvprsr.pvrCdeTxt = arune
			if atvxctr.hasCode {
				atvprsr.runesToParse[atvprsr.runesToParsei] = arune
				atvprsr.runesToParsei++
				if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
					captureActiveCode(curatvxctr, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
			} else {
				if strings.TrimSpace(string(arune)) != "" {
					if !atvxctr.foundCode {
						flushPassiveContent(curatvxctr, atvprsr, false)
						atvxctr.foundCode = true
					} else {
						flushPassiveContent(curatvxctr, atvprsr, false)
					}
					atvxctr.hasCode = true
					if len(atvprsr.runesToParse) == 0 {
						atvprsr.runesToParse = make([]rune, 81920)
					}
					atvprsr.runesToParse[atvprsr.runesToParsei] = arune
					atvprsr.runesToParsei++
					if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
						captureActiveCode(curatvxctr, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
						atvprsr.runesToParsei = 0
					}
				}
			}
		}
		atvxctr = nil
	}
	return
}

func setAtvA(atv *Active, d interface{}) {
	if atv.atvprsr != nil {
		if rdrRune, rdrRuneOk := d.(io.RuneReader); rdrRuneOk {
			atv.atvprsr.rdrRune = rdrRune
		}
		if rdrskr, rdrskrok := d.(io.Seeker); rdrskrok {
			atv.atvprsr.rdskr = rdrskr
		}
	}
	if prntr, prntrok := d.(iorw.Printing); prntrok {
		atv.printer = prntr
	}
	if atvmp, atvmpok := d.(map[string]interface{}); atvmpok {
		if len(atvmp) > 0 {
			for k, v := range atvmp {
				if len(atv.activeMap) == 0 {
					atv.activeMap = map[string]interface{}{}
				}
				atv.activeMap[k] = v
			}
		}
	}
}

func NewActive(maxBufSize int64, a ...interface{}) (atv *Active) {
	if maxBufSize < 81920 {
		maxBufSize = 81920
	}
	atv = &Active{atvprsr: &activeParser{
		maxBufSize: maxBufSize, lck: &sync.Mutex{},
		runesToParse:     make([]rune, maxBufSize),
		runeLabel:        [][]rune{[]rune("<@"), []rune("@>")},
		runeLabelI:       []int{0, 0},
		runesToParsei:    int(0),
		runePrvR:         []rune{rune(0)},
		psvLabel:         [][]rune{[]rune("<"), []rune(">")},
		psvLabelI:        []int{0, 0},
		psvPrvR:          []rune{rune(0)},
		psvRunesToParsei: int(0),
		psvRunesToParse:  make([]rune, maxBufSize),
		foundCdeTxt:      false,
		cdeTxt:           rune(0)}}

	atv.atvprsr.atv = atv

	for n, d := range a {
		if _, prntrok := d.(iorw.Printing); prntrok {
			setAtvA(atv, d)
			a = append(a[0:n], a[n+1:]...)
			break
		}
	}

	for n, d := range a {
		if callincl, callinclok := d.(func(...interface{}) interface{}); callinclok {
			if atv.callinclude == nil {
				atv.callinclude = callincl
				a = append(a[0:n], a[n+1:]...)
				break
			}
		}
	}

	for _, d := range a {
		setAtvA(atv, d)
	}
	return
}

func (atv *Active) Print(a ...interface{}) {
	if atv.printer != nil {
		atv.printer.Print(a...)
	}
}

func (atv *Active) Println(a ...interface{}) {
	if atv.printer != nil {
		atv.printer.Println(a...)
	}
}

func (atv *Active) Reset() {
	if atv.atvprsr != nil {
		atv.atvprsr.Reset()
	}
}

func (atv *Active) Close() {
	if atv.atvprsr != nil {
		atv.atvprsr.Close()
		atv.atvprsr = nil
	}

	if atv.printer != nil {
		atv.printer = nil
	}
	if atv.vm != nil {
		atv.vm = nil
	}
}

var activeGlobalMap = map[string]interface{}{}

func MapGlobal(name string, a interface{}) {
	if a != nil {
		if _, atvGlbOk := activeGlobalMap[name]; atvGlbOk {
			activeGlobalMap[name] = nil
		}
		activeGlobalMap[name] = a
	}
}

func MapGlobals(a ...interface{}) {
	if len(a) >= 2 {
		for len(a) > 0 {
			if s, sok := a[0].(string); sok {
				MapGlobal(s, a[1])
			}
			a[0] = nil
			a[1] = nil
			a = a[2:]
		}
	}
}
