package active

import (
	"fmt"
	goja "github.com/dop251/goja"
	gojaparse "github.com/dop251/goja/parser"
	iorw "github.com/efjoubert/lnksys/iorw"
	"io"
	"strings"
	"sync"
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
	atv                     *Active
	prgrm                   *goja.Program
}

func newActiveExecutor(atv *Active) (atvxctr *activeExecutor) {
	atvxctr = &activeExecutor{atv: atv, foundCode: false, hasCode: false, passiveBufferOffset: 0, lastPassiveBufferOffset: 0, activeBufferOffset: 0, lastActiveBufferOffset: 0}
	return
}

func (atvxctr *activeExecutor) passiveBuf() [][]rune {
	if atvxctr.passiveBuffer == nil {
		atvxctr.passiveBuffer = [][]rune{}
	}
	return atvxctr.passiveBuffer
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
	if atvxctr.atv != nil {
		atvxctr.atv = nil
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
	passiveRune  []rune
	passiveRunei int
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

	atvxctr []*activeExecutor
}

func (atvprsr *activeParser) atvxctor(prsnglvl int) (atvxctr *activeExecutor) {
	if atvprsr.atvxctr == nil {
		atvprsr.atvxctr = []*activeExecutor{}
	}
	if len(atvprsr.atvxctr) < prsnglvl+1 {
		atvprsr.atvxctr = append(atvprsr.atvxctr, newActiveExecutor(atvprsr.atv))
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

//func (atvprsr *activeParser) activeCode(atvcdelvl int) *iorw.BufferedRW {
//	return atvprsr.atvxctor(atvcdelvl).activeCode()
//}

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
}

func (atvprsr *activeParser) APrint(a ...interface{}) (err error) {
	if len(a) > 0 {
		atvprsr.lck.Lock()
		defer atvprsr.lck.Unlock()
		//atvprsr.atvbufrdr().Print(a...)
		var stopReading = false
		for _, d := range a {
			if rnrd, rnrdrok := d.(io.RuneReader); rnrdrok {
				if atvprsr.atvrdr != nil {
					for {
						if rne, rnsize, rnerr := atvprsr.atvrdr.ReadRune(); rnerr == nil {
							if rnsize > 0 {
								processRune(atvprsr.parsingLevel, rne, atvprsr, atvprsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
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
							processRune(atvprsr.parsingLevel, rne, atvprsr, atvprsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
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
						processRune(atvprsr.parsingLevel, rne, atvprsr, atvprsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
					}
				} else {
					if rnerr != io.EOF {
						err = rnerr
					}
					break
				}
			}
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
	flushPassiveContent(atvprsr.parsingLevel, atvprsr, true)
	if len(atvprsr.atvxctr) > atvprsr.parsingLevel {
		if atvprsr.atvxctr[atvprsr.parsingLevel].foundCode {
			flushActiveCode(atvprsr.parsingLevel, atvprsr, true)
		}
		atvxctr = atvprsr.atvxctr[atvprsr.parsingLevel]
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

func (atvprsr *activeParser) ACommit() (acerr error) {
	if atvprsr.atvrdr != nil {
		atvprsr.lck.Lock()
		defer func() {
			if err := recover(); err != nil {
				acerr = fmt.Errorf("Panic: %+v\n", err)
			}
			wrappingupActiveParsing(atvprsr)
			atvprsr.lck.Unlock()
		}()
		if atvxctr := preppingActiveParsing(atvprsr); atvxctr != nil && atvxctr.foundCode {
			if atvprsr.atv != nil {
				if atvprsr.atv.vm == nil {
					atvprsr.atv.vm = goja.New()
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
					atvxctr.PassivePrint(atvprsr.atv, fromOffset, toOffset)
				})
				if len(atvprsr.atv.activeMap) > 0 {
					for k, v := range atvprsr.atv.activeMap {
						if atvprsr.atv.vm.Get(k) != v {
							atvprsr.atv.vm.Set(k, v)
						}
					}
				}
				if len(activeGlobalMap) > 0 {
					for k, v := range activeGlobalMap {
						if atvprsr.atv.vm.Get(k) != v {
							atvprsr.atv.vm.Set(k, v)
						}
					}
				}
				var code = ""
				//if len(atvxctr.activeBuffer) > 0 {
				//	for _, atvb := range atvxctr.activeBuffer {
				//		code += string(atvb)
				//	}
				//}

				var nxtprm *goja.Program = nil
				var nxtprmerr error = nil
				pipeatvr, pipeatvw := io.Pipe()
				go func() {
					defer func() {
						pipeatvw.Close()
					}()
					for len(atvxctr.activeBuffer) > 0 {
						cde := string(atvxctr.activeBuffer[0])
						code += cde
						atvxctr.activeBuffer = atvxctr.activeBuffer[1:]
						iorw.FPrint(pipeatvw, cde)
					}
				}()

				var coderdr = pipeatvr //strings.NewReader(code)
				var parsedprgm, parsedprgmerr = gojaparse.ParseFile(nil, "", coderdr, 0)
				pipeatvr.Close()
				pipeatvr = nil
				pipeatvw = nil
				if parsedprgmerr == nil {
					nxtprm, nxtprmerr = goja.CompileAST(parsedprgm, false)
				} else {
					nxtprmerr = parsedprgmerr
					fmt.Println(nxtprmerr)
					fmt.Println(code)
					acerr = nxtprmerr
				}

				if acerr == nil && nxtprm != nil {
					var _, vmerr = atvprsr.atv.vm.RunProgram(nxtprm)
					if vmerr != nil {
						fmt.Println(vmerr)
						fmt.Println(code)
						acerr = vmerr
					}
				}
			}
		}
	}
	return
}

/*func commitActiveExecutor(atv *Active, atvxctr *activeExecutor) (acerr error) {
	if atv != nil {
		if atv.vm == nil {
			atv.vm = goja.New()
		}
		atv.vm.Set("out", atv)
		atv.vm.Set("CPrint", func(a ...interface{}) {
			cPrint(a...)
		})
		atv.vm.Set("CPrintln", func(a ...interface{}) {
			cPrint(a...)
			cPrint("\r\n")
		})
		atv.vm.Set("PassivePrint", func(fromOffset int64, toOffset int64) {
			atvxctr.PassivePrint(atv, fromOffset, toOffset)
		})
		if len(atv.activeMap) > 0 {
			for k, v := range atv.activeMap {
				if atv.vm.Get(k) != v {
					atv.vm.Set(k, v)
				}
			}
		}
		if len(activeGlobalMap) > 0 {
			for k, v := range activeGlobalMap {
				if atv.vm.Get(k) != v {
					atv.vm.Set(k, v)
				}
			}
		}
		var code = atvxctr.activeCode().String()
		var coderdr = strings.NewReader(code)
		var parsedprgm, parsedprgmerr = gojaparse.ParseFile(nil, "", coderdr, 0)
		if parsedprgmerr == nil {
			var prgm, prgmerr = goja.CompileAST(parsedprgm, false)
			if prgmerr == nil {
				var _, vmerr = atv.vm.RunProgram(prgm)
				if vmerr != nil {
					fmt.Println(vmerr)
					fmt.Println(code)
					acerr = vmerr
				}
			} else {
				fmt.Println(prgmerr)
				fmt.Println(code)
				acerr = prgmerr
			}
			prgm = nil
		} else {
			fmt.Println(parsedprgmerr)
			fmt.Println(code)
			acerr = parsedprgmerr
		}
		parsedprgm = nil
		atv.vm = nil
	}
	return acerr
}*/

func (atvprsr *activeParser) PassivePrint(psvbuflvl int, fromOffset int64, toOffset int64) {
	if len(atvprsr.atvxctr) > psvbuflvl {
		atvprsr.atvxctr[psvbuflvl].PassivePrint(atvprsr.atv, fromOffset, toOffset)
	}
}

func (atvprsr *activeParser) Print(a ...interface{}) {
	if atvprsr.atv != nil {
		atvprsr.atv.Print(a...)
	}
}

type Active struct {
	printer   iorw.Printing
	atvprsr   *activeParser
	vm        *goja.Runtime
	activeMap map[string]interface{}
}

func (atv *Active) APrint(a ...interface{}) (err error) {
	if atv.atvprsr != nil {
		err = atv.atvprsr.APrint(a...)
	}
	return
}

func (atv *Active) ACommit() (err error) {
	if atv.atvprsr != nil {
		err = atv.atvprsr.ACommit()
	}
	return
}

func (atv *Active) APrintln(a ...interface{}) {
	atv.APrint(a...)
	atv.APrint("/r/n")
}

func capturePassiveContent(psvcntlvl int, atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		if atvprsr.atvxctor(psvcntlvl).foundCode {
			if len(atvprsr.passiveRune) == 0 {
				atvprsr.passiveRune = make([]rune, 81920)
			}
			if n < pl && atvprsr.passiveRunei < len(atvprsr.passiveRune) {
				if (pl - n) >= (len(atvprsr.passiveRune) - atvprsr.passiveRunei) {
					var cl = copy(atvprsr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)], p[n:n+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)])
					atvprsr.passiveRunei += cl
					n += cl
					atvprsr.atvxctor(psvcntlvl).passiveBufferOffset += int64(cl)
				} else if (pl - n) < (len(atvprsr.passiveRune) - atvprsr.passiveRunei) {
					var cl = copy(atvprsr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(pl-n)], p[n:n+(pl-n)])
					atvprsr.passiveRunei += cl
					n += cl
					atvprsr.atvxctor(psvcntlvl).passiveBufferOffset += int64(cl)
				}
				if len(atvprsr.passiveRune) == atvprsr.passiveRunei {
					var psvRunes = make([]rune, atvprsr.passiveRunei)
					copy(psvRunes, atvprsr.passiveRune[0:atvprsr.passiveRunei])
					atvprsr.atvxctor(psvcntlvl).passiveBuf()
					atvprsr.atvxctor(psvcntlvl).passiveBuffer = append(atvprsr.atvxctor(psvcntlvl).passiveBuffer, psvRunes)
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
	return
}

func flushPassiveContent(psvlvl int, atvprsr *activeParser, force bool) {
	if atvprsr.runesToParsei > 0 {
		processUnparsedPassiveContent(psvlvl, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
		atvprsr.runesToParsei = 0
	}

	if atvprsr.psvRunesToParsei > 0 {
		capturePassiveContent(psvlvl, atvprsr, atvprsr.psvRunesToParse[0:atvprsr.psvRunesToParsei])
		atvprsr.psvRunesToParsei = 0
	}
	if atvprsr.atvxctor(psvlvl).foundCode {
		if force {
			if atvprsr.passiveRunei > 0 {
				var psvRunes = make([]rune, atvprsr.passiveRunei)
				copy(psvRunes, atvprsr.passiveRune[0:atvprsr.passiveRunei])
				atvprsr.atvxctor(psvlvl).passiveBuf()
				atvprsr.atvxctor(psvlvl).passiveBuffer = append(atvprsr.atvxctor(atvprsr.parsingLevel).passiveBuffer, psvRunes)
				psvRunes = nil
				atvprsr.passiveRunei = 0
			}
		}

		if atvprsr.atvxctor(psvlvl).lastPassiveBufferOffset < atvprsr.atvxctor(psvlvl).passiveBufferOffset {
			for _, arune := range []rune(fmt.Sprintf("PassivePrint(%d,%d);", atvprsr.atvxctor(psvlvl).lastPassiveBufferOffset, atvprsr.atvxctor(psvlvl).passiveBufferOffset)) {
				if len(atvprsr.runesToParse) == 0 {
					atvprsr.runesToParse = make([]rune, 81920)
				}
				atvprsr.runesToParse[atvprsr.runesToParsei] = arune
				atvprsr.runesToParsei++
				if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
					captureActiveCode(psvlvl, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
			}
			atvprsr.atvxctor(psvlvl).lastPassiveBufferOffset = atvprsr.atvxctor(psvlvl).passiveBufferOffset
		}
	}
}

func (atv *Active) PassivePrint(psvbuflvl int, fromOffset int64, toOffset int64) {
	if atv.atvprsr != nil {
		atv.atvprsr.PassivePrint(psvbuflvl, fromOffset, toOffset)
	}
}

func processUnparsedPassiveContent(psvlvl int, atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	if pl > 0 {
		flushActiveCode(psvlvl, atvprsr, false)
	}
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
			capturePassiveContent(psvlvl, atvprsr, atvprsr.psvRunesToParse[0:atvprsr.psvRunesToParsei])
			atvprsr.psvRunesToParsei = 0
		}
	}
	return
}

func processRune(processlvl int, rne rune, atvprsr *activeParser, runelbl [][]rune, runelbli []int, runePrvR []rune) {
	if runelbli[1] == 0 && runelbli[0] < len(runelbl[0]) {
		if runelbli[0] > 0 && runelbl[0][runelbli[0]-1] == runePrvR[0] && runelbl[0][runelbli[0]] != rne {
			processUnparsedPassiveContent(processlvl, atvprsr, runelbl[0][0:runelbli[0]])
			runelbli[0] = 0
			runePrvR[0] = rune(0)
		}
		if runelbl[0][runelbli[0]] == rne {
			runelbli[0]++
			if len(runelbl[0]) == runelbli[0] {
				atvprsr.atvxctor(processlvl).hasCode = false
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[0] > 0 {
				processUnparsedPassiveContent(processlvl, atvprsr, runelbl[0][0:runelbli[0]])
				runelbli[0] = 0
			}
			runePrvR[0] = rne
			processUnparsedPassiveContent(processlvl, atvprsr, runePrvR)
		}
	} else if runelbli[0] == len(runelbl[0]) && runelbli[1] < len(runelbl[1]) {
		if runelbli[1] > 0 && runelbl[1][runelbli[1]-1] == runePrvR[0] && runelbl[1][runelbli[1]] != rne {
			processUnparsedActiveCode(processlvl, atvprsr, runelbl[1][0:runelbli[1]])
			runelbli[1] = 0
			runePrvR[0] = rune(0)
		}
		if runelbl[1][runelbli[1]] == rne {
			runelbli[1]++
			if runelbli[1] == len(runelbl[1]) {
				if atvprsr.runesToParsei > 0 {
					captureActiveCode(processlvl, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
				runePrvR[0] = rune(0)
				runelbli[0] = 0
				runelbli[1] = 0
				atvprsr.atvxctor(processlvl).hasCode = false
				atvprsr.atvxctor(processlvl).lastPassiveBufferOffset = atvprsr.atvxctor(processlvl).passiveBufferOffset
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[1] > 0 {
				processUnparsedActiveCode(processlvl, atvprsr, runelbl[1][0:runelbli[1]])
				runelbli[1] = 0
			}
			runePrvR[0] = rne
			processUnparsedActiveCode(processlvl, atvprsr, runePrvR)
		}
	}
}

func captureActiveCode(atvcdelvl int, atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		if len(atvprsr.activeRune) == 0 {
			atvprsr.activeRune = make([]rune, 81920)
		}
		if n < pl && atvprsr.activeRunei < len(atvprsr.activeRune) {
			if (pl - n) >= (len(atvprsr.activeRune) - atvprsr.activeRunei) {
				var cl = copy(atvprsr.activeRune[atvprsr.activeRunei:atvprsr.activeRunei+(len(atvprsr.activeRune)-atvprsr.activeRunei)], p[n:n+(len(atvprsr.activeRune)-atvprsr.activeRunei)])
				atvprsr.activeRunei += cl
				n += cl
				atvprsr.atvxctor(atvcdelvl).activeBufferOffset += int64(cl)
			} else if (pl - n) < (len(atvprsr.activeRune) - atvprsr.activeRunei) {
				var cl = copy(atvprsr.activeRune[atvprsr.activeRunei:atvprsr.activeRunei+(pl-n)], p[n:n+(pl-n)])
				atvprsr.activeRunei += cl
				n += cl
				atvprsr.atvxctor(atvcdelvl).activeBufferOffset += int64(cl)
			}
			if len(atvprsr.activeRune) == atvprsr.activeRunei {
				var atvRunes = make([]rune, atvprsr.activeRunei)
				copy(atvRunes, atvprsr.activeRune[0:atvprsr.activeRunei])
				atvprsr.atvxctor(atvcdelvl).activeBuf()
				atvprsr.atvxctor(atvcdelvl).activeBuffer = append(atvprsr.atvxctor(atvcdelvl).activeBuffer, atvRunes)
				atvRunes = nil
				atvprsr.activeRunei = 0
			}
		} else {
			break
		}
	}
	return
}

func flushActiveCode(atvcdelvl int, atvprsr *activeParser, force bool) {
	if atvprsr.runesToParsei > 0 {
		captureActiveCode(atvcdelvl, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
		atvprsr.runesToParsei = 0
	}
	if force {
		if atvprsr.activeRunei > 0 {
			var atvRunes = make([]rune, atvprsr.activeRunei)
			copy(atvRunes, atvprsr.activeRune[0:atvprsr.activeRunei])
			atvprsr.atvxctor(atvcdelvl).activeBuf()
			atvprsr.atvxctor(atvcdelvl).activeBuffer = append(atvprsr.atvxctor(atvprsr.parsingLevel).activeBuffer, atvRunes)
			atvRunes = nil
			atvprsr.activeRunei = 0
		}
	}
}

func processUnparsedActiveCode(processlvl int, atvprsr *activeParser, p []rune) (err error) {
	if len(p) > 0 {
		for _, arune := range p {
			if atvprsr.atvxctor(processlvl).hasCode {
				atvprsr.runesToParse[atvprsr.runesToParsei] = arune
				atvprsr.runesToParsei++
				if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
					captureActiveCode(processlvl, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
			} else {
				if strings.TrimSpace(string(arune)) != "" {
					if !atvprsr.atvxctor(processlvl).foundCode {
						flushPassiveContent(processlvl, atvprsr, false)
						atvprsr.atvxctor(processlvl).foundCode = true
					} else {
						flushPassiveContent(processlvl, atvprsr, false)
					}
					atvprsr.atvxctor(processlvl).hasCode = true
					if len(atvprsr.runesToParse) == 0 {
						atvprsr.runesToParse = make([]rune, 81920)
					}
					atvprsr.runesToParse[atvprsr.runesToParsei] = arune
					atvprsr.runesToParsei++
					if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
						captureActiveCode(processlvl, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
						atvprsr.runesToParsei = 0
					}
				}
			}
		}
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
		psvRunesToParse:  make([]rune, maxBufSize)}}

	atv.atvprsr.atv = atv

	for n, d := range a {
		if _, prntrok := d.(iorw.Printing); prntrok {
			setAtvA(atv, d)
			a = append(a[0:n], a[n+1:]...)
			break
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
