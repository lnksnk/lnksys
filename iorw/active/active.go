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

type activeParser struct {
	atv        *Active
	atvrdr     *iorw.BufferedRW
	rdrRune    io.RuneReader
	rdskr      io.Seeker
	maxBufSize int64
	lck        *sync.RWMutex
	//
	runesToParse  []rune
	runesToParsei int
	runeLabel     [][]rune
	runeLabelI    []int
	runePrvR      []rune
	//
	passiveRune             []rune
	passiveRunei            int
	passiveBuffer           [][][]rune
	passiveBufferi          int
	passiveBufferOffset     int64
	lastPassiveBufferOffset int64
	//
	psvRunesToParse  []rune
	psvRunesToParsei int
	psvLabel         [][]rune
	psvLabelI        []int
	psvPrvR          []rune
	//
	hasCode   bool
	foundCode bool

	curAtvCde []*iorw.BufferedRW
}

func (atvprsr *activeParser) atvbufrdr() *iorw.BufferedRW {
	if atvprsr.atvrdr == nil {
		atvprsr.atvrdr = iorw.NewBufferedRW(atvprsr.maxBufSize, nil)
	}
	return atvprsr.atvrdr
}

func (atvprsr *activeParser) activeCode(atvcdelvl int) *iorw.BufferedRW {
	if atvprsr.curAtvCde == nil {
		atvprsr.curAtvCde = []*iorw.BufferedRW{}
	}
	if len(atvprsr.curAtvCde) < atvcdelvl+1 {
		atvprsr.curAtvCde = append(atvprsr.curAtvCde)
	}
	return atvprsr.curAtvCde[atvcdelvl]
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
	if atvprsr.passiveBuffer != nil {
		for len(atvprsr.passiveBuffer) > 0 {
			atvprsr.passiveBuffer[0] = nil
			atvprsr.passiveBuffer = atvprsr.passiveBuffer[1:]
		}
		atvprsr.passiveBuffer = nil
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
	if atvprsr.curAtvCde != nil {
		for len(atvprsr.curAtvCde) > 0 {
			atvprsr.curAtvCde[0].Close()
			atvprsr.curAtvCde[0] = nil
			atvprsr.curAtvCde = atvprsr.curAtvCde[1:]
		}
		atvprsr.curAtvCde = nil
	}
	if atvprsr.atv != nil {
		atvprsr.atv = nil
	}
}

func (atvprsr *activeParser) APrint(a ...interface{}) (err error) {
	atvprsr.lck.RLock()
	defer atvprsr.lck.RUnlock()
	atvprsr.atvbufrdr().Print(a...)
	for {
		if rne, rnsize, rnerr := atvprsr.atvrdr.ReadRune(); rnerr == nil {
			if rnsize > 0 {
				processRune(rne, atvprsr, atvprsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
			}
		} else {
			if rnerr != io.EOF {
				err = rnerr
			}
			break
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

func preppingActiveParsing(atvprsr *activeParser) {
	flushPassiveContent(atvprsr.passiveBufferi, atvprsr, true)
	atvprsr.passiveBufferi++
	if atvprsr.foundCode {
		flushActiveCode(atvprsr.passiveBufferi-1, atvprsr)
	} else {

	}
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
	if atvprsr.passiveRunei > 0 {
		atvprsr.passiveRunei = 0
	}
	if len(atvprsr.psvLabelI) == 2 {
		atvprsr.psvLabelI[0] = 0
		atvprsr.psvLabelI[1] = 0
	}
	if len(atvprsr.psvPrvR) == 1 {
		atvprsr.psvPrvR[0] = 0
	}
}

func wrappingupActiveParsing(atvprsr *activeParser) {
	if atvprsr.passiveBufferi > 0 {
		atvprsr.passiveBufferi--
		for len(atvprsr.passiveBuffer) > atvprsr.passiveBufferi {
			var psvbufi = len(atvprsr.passiveBuffer) - 1
			for len(atvprsr.passiveBuffer[psvbufi]) > 0 {
				atvprsr.passiveBuffer[psvbufi][0] = nil
				atvprsr.passiveBuffer[psvbufi] = atvprsr.passiveBuffer[psvbufi][1:]
			}
			atvprsr.passiveBuffer[atvprsr.passiveBufferi] = nil
			if atvprsr.passiveBufferi > 0 {
				atvprsr.passiveBuffer = atvprsr.passiveBuffer[:atvprsr.passiveBufferi]
			} else {
				atvprsr.passiveBuffer = nil
			}
		}

		if len(atvprsr.curAtvCde) > 0 {
			for len(atvprsr.passiveBuffer) > atvprsr.passiveBufferi {
				var atvbufi = len(atvprsr.passiveBuffer) - 1
				atvprsr.curAtvCde[atvbufi].Close()
				atvprsr.curAtvCde[atvbufi] = nil
				if atvprsr.passiveBufferi > 0 {
					atvprsr.curAtvCde = atvprsr.curAtvCde[:atvprsr.passiveBufferi]
				} else {
					atvprsr.curAtvCde = nil
				}
			}
		}
	}
}

func (atvprsr *activeParser) ACommit() (acerr error) {
	if atvprsr.atvrdr != nil {
		atvprsr.lck.RLock()
		defer func() {
			wrappingupActiveParsing(atvprsr)
			atvprsr.lck.RUnlock()
		}()
		preppingActiveParsing(atvprsr)
		if atvprsr.foundCode {
			func() {
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
					atvprsr.atv.vm.Set("_atvprsr", atvprsr)
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
					var code = atvprsr.activeCode(atvprsr.passiveBufferi - 1).String()
					var coderdr = strings.NewReader(code)
					var parsedprgm, parsedprgmerr = gojaparse.ParseFile(nil, "", coderdr, 0) //goja.Compile("", code, false)
					if parsedprgmerr == nil {
						var prgm, prgmerr = goja.CompileAST(parsedprgm, false)
						if prgmerr == nil {
							var _, vmerr = atvprsr.atv.vm.RunProgram(prgm)
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
					atvprsr.atv.vm = nil
				}
			}()
		}
	}
	return
}

func (atvprsr *activeParser) PassivePrint(psvbuflvl int, fromOffset int64, toOffset int64) {

	if len(atvprsr.passiveBuffer) > 0 {
		if fromOffset >= 0 && toOffset <= atvprsr.passiveBufferOffset {
			var psi = int(0)
			var pei = int(0)
			var pfrom = int64(0)
			var pto = int64(0)
			var pl = int64(0)
			for _, psvb := range atvprsr.passiveBuffer {
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
						atvprsr.atv.Print(string(psvb[psvbuflvl][psi:pei]))
						if pto == toOffset {
							break
						}
					} else if toOffset < pto {
						if pto-toOffset > 0 {
							pei = int(pl - (pto - toOffset))
							atvprsr.atv.Print(string(psvb[psvbuflvl][psi:pei]))
						}
						break
					}
				}
				pfrom += pto
			}
		}
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

type activeRune struct {
	rne     rune
	rnesize int
	reerr   error
	atvprsr *activeParser
}

func (atvRne *activeRune) process() {
	processRune(atvRne.rne, atvRne.atvprsr, atvRne.atvprsr.runeLabel, atvRne.atvprsr.runeLabelI, atvRne.atvprsr.runePrvR)
}

func (atvRne *activeRune) close() {
	atvRne.atvprsr = nil
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

func capturePassiveContent(atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		if atvprsr.foundCode {
			if len(atvprsr.passiveRune) == 0 {
				atvprsr.passiveRune = make([]rune, 81920)
			}
			if n < pl && atvprsr.passiveRunei < len(atvprsr.passiveRune) {
				if (pl - n) >= (len(atvprsr.passiveRune) - atvprsr.passiveRunei) {
					var cl = copy(atvprsr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)], p[n:n+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)])
					atvprsr.passiveRunei += cl
					n += cl
					atvprsr.passiveBufferOffset += int64(cl)
				} else if (pl - n) < (len(atvprsr.passiveRune) - atvprsr.passiveRunei) {
					var cl = copy(atvprsr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(pl-n)], p[n:n+(pl-n)])
					atvprsr.passiveRunei += cl
					n += cl
					atvprsr.passiveBufferOffset += int64(cl)
				}
				if len(atvprsr.passiveRune) == atvprsr.passiveRunei {
					var psvRunes = make([]rune, atvprsr.passiveRunei)
					copy(psvRunes, atvprsr.passiveRune[0:atvprsr.passiveRunei])
					if len(atvprsr.passiveBuffer) == 0 {
						atvprsr.passiveBuffer = [][][]rune{}
					}
					if len(atvprsr.passiveBuffer) < atvprsr.passiveBufferi {
						atvprsr.passiveBuffer = append(atvprsr.passiveBuffer, [][]rune{})
					}
					atvprsr.passiveBuffer[atvprsr.passiveBufferi] = append(atvprsr.passiveBuffer[atvprsr.passiveBufferi], psvRunes)
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
		capturePassiveContent(atvprsr, atvprsr.psvRunesToParse[0:atvprsr.psvRunesToParsei])
		atvprsr.psvRunesToParsei = 0
	}
	if atvprsr.foundCode {
		if force {
			if atvprsr.passiveRunei > 0 {
				var psvRunes = make([]rune, atvprsr.passiveRunei)
				copy(psvRunes, atvprsr.passiveRune[0:atvprsr.passiveRunei])
				if len(atvprsr.passiveBuffer) == 0 {
					atvprsr.passiveBuffer = [][][]rune{}
				}
				if len(atvprsr.passiveBuffer) < atvprsr.passiveBufferi {
					atvprsr.passiveBuffer = append(atvprsr.passiveBuffer, [][]rune{})
				}
				atvprsr.passiveBuffer[atvprsr.passiveBufferi] = append(atvprsr.passiveBuffer[atvprsr.passiveBufferi], psvRunes)
				psvRunes = nil
				atvprsr.passiveRunei = 0
			}
		}

		if atvprsr.lastPassiveBufferOffset < atvprsr.passiveBufferOffset {
			for _, arune := range []rune(fmt.Sprintf("_atvprsr.PassivePrint(%d,%d,%d);", atvprsr.passiveBufferi, atvprsr.lastPassiveBufferOffset, atvprsr.passiveBufferOffset)) {
				if len(atvprsr.runesToParse) == 0 {
					atvprsr.runesToParse = make([]rune, 81920)
				}
				atvprsr.runesToParse[atvprsr.runesToParsei] = arune
				atvprsr.runesToParsei++
				if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
					captureActiveCode(atvprsr.passiveBufferi, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
			}
			atvprsr.lastPassiveBufferOffset = atvprsr.passiveBufferOffset
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
		flushActiveCode(psvlvl, atvprsr)
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
			capturePassiveContent(atvprsr, atvprsr.psvRunesToParse[0:atvprsr.psvRunesToParsei])
			atvprsr.psvRunesToParsei = 0
		}
	}
	return
}

func processRune(rne rune, atvprsr *activeParser, runelbl [][]rune, runelbli []int, runePrvR []rune) {
	if runelbli[1] == 0 && runelbli[0] < len(runelbl[0]) {
		if runelbli[0] > 0 && runelbl[0][runelbli[0]-1] == runePrvR[0] && runelbl[0][runelbli[0]] != rne {
			processUnparsedPassiveContent(atvprsr.passiveBufferi, atvprsr, runelbl[0][0:runelbli[0]])
			runelbli[0] = 0
			runePrvR[0] = rune(0)
		}
		if runelbl[0][runelbli[0]] == rne {
			runelbli[0]++
			if len(runelbl[0]) == runelbli[0] {
				atvprsr.hasCode = false
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[0] > 0 {
				processUnparsedPassiveContent(atvprsr.passiveBufferi, atvprsr, runelbl[0][0:runelbli[0]])
				runelbli[0] = 0
			}
			runePrvR[0] = rne
			processUnparsedPassiveContent(atvprsr.passiveBufferi, atvprsr, runePrvR)
		}
	} else if runelbli[0] == len(runelbl[0]) && runelbli[1] < len(runelbl[1]) {
		if runelbli[1] > 0 && runelbl[1][runelbli[1]-1] == runePrvR[0] && runelbl[1][runelbli[1]] != rne {
			processUnparsedActiveCode(atvprsr, runelbl[1][0:runelbli[1]])
			runelbli[1] = 0
			runePrvR[0] = rune(0)
		}
		if runelbl[1][runelbli[1]] == rne {
			runelbli[1]++
			if runelbli[1] == len(runelbl[1]) {
				if atvprsr.runesToParsei > 0 {
					captureActiveCode(atvprsr.passiveBufferi, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
				runePrvR[0] = rune(0)
				runelbli[0] = 0
				runelbli[1] = 0
				atvprsr.hasCode = false
				atvprsr.lastPassiveBufferOffset = atvprsr.passiveBufferOffset
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[1] > 0 {
				processUnparsedActiveCode(atvprsr, runelbl[1][0:runelbli[1]])
				runelbli[1] = 0
			}
			runePrvR[0] = rne
			processUnparsedActiveCode(atvprsr, runePrvR)
		}
	}
}

func captureActiveCode(atvcdelvl int, atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		atvprsr.activeCode(atvcdelvl).Print(string(p[n : n+(pl-n)]))
		n += pl
	}
	return
}

func flushActiveCode(atvcdelvl int, atvprsr *activeParser) {
	if atvprsr.runesToParsei > 0 {
		captureActiveCode(atvcdelvl, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
		atvprsr.runesToParsei = 0
	}
}

func processUnparsedActiveCode(atvprsr *activeParser, p []rune) (err error) {
	if len(p) > 0 {
		for _, arune := range p {
			if atvprsr.hasCode {
				atvprsr.runesToParse[atvprsr.runesToParsei] = arune
				atvprsr.runesToParsei++
				if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
					captureActiveCode(atvprsr.passiveBufferi, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
					atvprsr.runesToParsei = 0
				}
			} else {
				if strings.TrimSpace(string(arune)) != "" {
					if !atvprsr.foundCode {
						flushPassiveContent(atvprsr.passiveBufferi, atvprsr, false)
						atvprsr.foundCode = true
					} else {
						flushPassiveContent(atvprsr.passiveBufferi, atvprsr, false)
					}
					atvprsr.hasCode = true
					if len(atvprsr.runesToParse) == 0 {
						atvprsr.runesToParse = make([]rune, 81920)
					}
					atvprsr.runesToParse[atvprsr.runesToParsei] = arune
					atvprsr.runesToParsei++
					if atvprsr.runesToParsei == len(atvprsr.runesToParse) {
						captureActiveCode(atvprsr.passiveBufferi, atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
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
		maxBufSize: maxBufSize, lck: &sync.RWMutex{},
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
