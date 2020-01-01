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

type Active struct {
	lck        *sync.RWMutex
	maxBufSize int64
	rdrRune    io.RuneReader
	rdskr      io.Seeker
	printer    iorw.Printing
	//
	runesToParse  []rune
	runesToParsei int
	runeLabel     [][]rune
	runeLabelI    []int
	runePrvR      []rune
	//
	passiveRune             []rune
	passiveRunei            int
	passiveBuffer           [][]rune
	passiveBufferOffset     int64
	lastPassiveBufferOffset int64
	//
	psvRunesToParse  []rune
	psvRunesToParsei int
	psvLabel         [][]rune
	psvLabelI        []int
	psvPrvR          []rune
	//
	hasCode          bool
	foundCode        bool
	atvRunesToParse  []rune
	atvRunesToParsei int

	curAtvCde *iorw.BufferedRW
	vm        *goja.Runtime
	activeMap map[string]interface{}
}

func (atv *Active) activeCode() *iorw.BufferedRW {
	if atv.curAtvCde == nil {
		atv.curAtvCde = iorw.NewBufferedRW(atv.maxBufSize, nil)
	}
	return atv.curAtvCde
}

type activeRune struct {
	rne     rune
	rnesize int
	rneerr  error
	atv     *Active
}

func (atvRne *activeRune) process() {
	processRune(atvRne.rne, atvRne.atv, atvRne.atv.runeLabel, atvRne.atv.runeLabelI, atvRne.atv.runePrvR)
}

func (atvRne *activeRune) close() {
	atvRne.atv = nil
}

func (atv *Active) ExecuteActive(maxbufsize int) (atverr error) {
	atv.lck.RLock()
	defer atv.lck.RUnlock()
	if len(atv.runesToParse) == 0 {
		atv.runesToParse = make([]rune, maxbufsize)
	}
	atv.runesToParsei = int(0)
	var atvCntntRunesErr = error(nil)
	if len(atv.runeLabel) == 0 {
		atv.runeLabel = [][]rune{[]rune("<@"), []rune("@>")}
		atv.runeLabelI = []int{0, 0}
		if len(atv.runePrvR) == 0 {
			atv.runePrvR = []rune{rune(0)}
		}
		atv.runePrvR[0] = rune(0)
	}
	if len(atv.psvLabel) == 0 {
		atv.psvLabel = [][]rune{[]rune("<"), []rune(">")}
		atv.psvLabelI = []int{0, 0}
		if len(atv.psvPrvR) == 0 {
			atv.psvPrvR = []rune{rune(0)}
		}
		atv.psvPrvR[0] = rune(0)
	}
	if atv.rdrRune != nil {
		func() bool{
			var doneActRead = make(chan bool,1)
			defer close(doneActRead)
			go func() {
				for atvCntntRunesErr == nil {
					if rne, rnsize, rnerr := atv.rdrRune.ReadRune(); rnerr == nil {
						if rnsize > 0 {
							processRune(rne, atv, atv.runeLabel, atv.runeLabelI, atv.runePrvR)
						}
					} else {
						if rnerr != io.EOF {
							atverr = rnerr
						}
						break
					}
				}
				doneActRead<-true
			}()
			return <-doneActRead
		}()
		if atverr == nil {
			flushPassiveContent(atv, true)
			if atv.foundCode {
				flushActiveCode(atv)
				func() {
					if atv.vm == nil {
						atv.vm = goja.New()
					}
					atv.vm.Set("out", atv)
					atv.vm.Set("_atv", atv)
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
					var code = atv.activeCode()
					var parsedprgm, parsedprgmerr = gojaparse.ParseFile(nil, "", code, 0) //goja.Compile("", code, false)
					if parsedprgmerr == nil {
						var prgm, prgmerr = goja.CompileAST(parsedprgm, false)
						if prgmerr == nil {
							var _, vmerr = atv.vm.RunProgram(prgm)
							if vmerr != nil {
								fmt.Println(vmerr)
								//fmt.Println(code)
								atverr = vmerr
							}
						} else {
							fmt.Println(prgmerr)
							//fmt.Println(code)
							atverr = prgmerr
						}
						prgm = nil
					} else {
						fmt.Println(parsedprgmerr)
						//fmt.Println(code)
						atverr = parsedprgmerr
					}
					parsedprgm = nil
					atv.vm = nil
				}()
			}
		}
	}
	return
}

func capturePassiveContent(atv *Active, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		if atv.foundCode {
			if len(atv.passiveRune) == 0 {
				atv.passiveRune = make([]rune, 81920)
			}
			if n < pl && atv.passiveRunei < len(atv.passiveRune) {
				if (pl - n) >= (len(atv.passiveRune) - atv.passiveRunei) {
					var cl = copy(atv.passiveRune[atv.passiveRunei:atv.passiveRunei+(len(atv.passiveRune)-atv.passiveRunei)], p[n:n+(len(atv.passiveRune)-atv.passiveRunei)])
					atv.passiveRunei += cl
					n += cl
					atv.passiveBufferOffset += int64(cl)
				} else if (pl - n) < (len(atv.passiveRune) - atv.passiveRunei) {
					var cl = copy(atv.passiveRune[atv.passiveRunei:atv.passiveRunei+(pl-n)], p[n:n+(pl-n)])
					atv.passiveRunei += cl
					n += cl
					atv.passiveBufferOffset += int64(cl)
				}
				if len(atv.passiveRune) == atv.passiveRunei {
					var psvRunes = make([]rune, atv.passiveRunei)
					copy(psvRunes, atv.passiveRune[0:atv.passiveRunei])
					if len(atv.passiveBuffer) == 0 {
						atv.passiveBuffer = [][]rune{}
					}
					atv.passiveBuffer = append(atv.passiveBuffer, psvRunes)
					psvRunes = nil
					atv.passiveRunei = 0
				}
			} else {
				break
			}
		} else {
			atv.Print(string(p))
			n += pl
		}
	}
	return
}

func flushPassiveContent(atv *Active, force bool) {
	if atv.runesToParsei > 0 {
		capturePassiveContent(atv, atv.runesToParse[0:atv.runesToParsei])
		atv.runesToParsei = 0
	}

	if atv.psvRunesToParsei > 0 {
		capturePassiveContent(atv, atv.psvRunesToParse[0:atv.psvRunesToParsei])
		atv.psvRunesToParsei = 0
	}
	if atv.foundCode {
		if force {
			if atv.passiveRunei > 0 {
				var psvRunes = make([]rune, atv.passiveRunei)
				copy(psvRunes, atv.passiveRune[0:atv.passiveRunei])
				if len(atv.passiveBuffer) == 0 {
					atv.passiveBuffer = [][]rune{}
				}
				atv.passiveBuffer = append(atv.passiveBuffer, psvRunes)
				psvRunes = nil
				atv.passiveRunei = 0
			}
		}

		if atv.lastPassiveBufferOffset < atv.passiveBufferOffset {
			for _, arune := range []rune(fmt.Sprintf("_atv.PassivePrint(%d,%d);", atv.lastPassiveBufferOffset, atv.passiveBufferOffset)) {
				if len(atv.atvRunesToParse) == 0 {
					atv.atvRunesToParse = make([]rune, 81920)
				}
				atv.atvRunesToParse[atv.atvRunesToParsei] = arune
				atv.atvRunesToParsei++
				if atv.atvRunesToParsei == len(atv.atvRunesToParse) {
					captureActiveCode(atv, atv.atvRunesToParse[0:atv.atvRunesToParsei])
					atv.atvRunesToParsei = 0
				}
			}
			atv.lastPassiveBufferOffset = atv.passiveBufferOffset
		}
	}
}

func (atv *Active) PassivePrint(fromOffset int64, toOffset int64) {
	if len(atv.passiveBuffer) > 0 {
		if fromOffset >= 0 && toOffset <= atv.passiveBufferOffset {
			var psi = int(0)
			var pei = int(0)
			var pfrom = int64(0)
			var pto = int64(0)
			var pl = int64(0)
			for _, psvb := range atv.passiveBuffer {
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
						atv.Print(string(psvb[psi:pei]))
						if pto == toOffset {
							break
						}
					} else if toOffset < pto {
						if pto-toOffset > 0 {
							pei = int(pl - (pto - toOffset))
							atv.Print(string(psvb[psi:pei]))
						}
						break
					}
				}
				pfrom += pto
			}
		}
	}
}

func processUnparsedPassiveContent(atv *Active, p []rune) (n int, err error) {
	var pl = len(p)
	if pl > 0 {
		flushActiveCode(atv)
	}
	for n < pl && atv.runesToParsei < len(atv.runesToParse) {
		if (pl - n) >= (len(atv.runesToParse) - atv.runesToParsei) {

			var cl = copy(atv.runesToParse[atv.runesToParsei:atv.runesToParsei+(len(atv.runesToParse)-atv.runesToParsei)], p[n:n+(len(atv.runesToParse)-atv.runesToParsei)])
			n += cl
			atv.runesToParsei += cl
		} else if (pl - n) < (len(atv.runesToParse) - atv.runesToParsei) {
			var cl = copy(atv.runesToParse[atv.runesToParsei:atv.runesToParsei+(pl-n)], p[n:n+(pl-n)])
			n += cl
			atv.runesToParsei += cl
		}
		if atv.runesToParsei > 0 && atv.runesToParsei == len(atv.runesToParse) {
			capturePassiveContent(atv, atv.runesToParse[0:atv.runesToParsei])
			atv.runesToParsei = 0
		}
	}
	return
}

func processRune(rne rune, atv *Active, runelbl [][]rune, runelbli []int, runePrvR []rune) {
	if runelbli[1] == 0 && runelbli[0] < len(runelbl[0]) {
		if runelbli[0] > 0 && runelbl[0][runelbli[0]-1] == runePrvR[0] && runelbl[0][runelbli[0]] != rne {
			processUnparsedPassiveContent(atv, runelbl[0][0:runelbli[0]])
			runelbli[0] = 0
			runePrvR[0] = rune(0)
		}
		if runelbl[0][runelbli[0]] == rne {
			runelbli[0]++
			if len(runelbl[0]) == runelbli[0] {
				atv.hasCode = false
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[0] > 0 {
				processUnparsedPassiveContent(atv, runelbl[0][0:runelbli[0]])
				runelbli[0] = 0
			}
			runePrvR[0] = rne
			processUnparsedPassiveContent(atv, runePrvR)
		}
	} else if runelbli[0] == len(runelbl[0]) && runelbli[1] < len(runelbl[1]) {
		if runelbli[1] > 0 && runelbl[1][runelbli[1]-1] == runePrvR[0] && runelbl[1][runelbli[1]] != rne {
			processUnparsedActiveCode(atv, runelbl[1][0:runelbli[1]])
			runelbli[1] = 0
			runePrvR[0] = rune(0)
		}
		if runelbl[1][runelbli[1]] == rne {
			runelbli[1]++
			if runelbli[1] == len(runelbl[1]) {
				if atv.atvRunesToParsei > 0 {
					captureActiveCode(atv, atv.atvRunesToParse[0:atv.atvRunesToParsei])
					atv.atvRunesToParsei = 0
				}
				runePrvR[0] = rune(0)
				runelbli[0] = 0
				runelbli[1] = 0
				atv.hasCode = false
				atv.lastPassiveBufferOffset = atv.passiveBufferOffset
			} else {
				runePrvR[0] = rne
			}
		} else {
			if runelbli[1] > 0 {
				processUnparsedActiveCode(atv, runelbl[1][0:runelbli[1]])
				runelbli[1] = 0
			}
			runePrvR[0] = rne
			processUnparsedActiveCode(atv, runePrvR)
		}
	}
}

func captureActiveCode(atv *Active, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		atv.activeCode().Print(string(p[n : n+(pl-n)]))
		n += pl
	}
	return
}

func flushActiveCode(atv *Active) {
	if atv.atvRunesToParsei > 0 {
		captureActiveCode(atv, atv.atvRunesToParse[0:atv.atvRunesToParsei])
		atv.atvRunesToParsei = 0
	}
}

func processUnparsedActiveCode(atv *Active, p []rune) (err error) {
	if len(p) > 0 {
		for _, arune := range p {
			if atv.hasCode {
				atv.atvRunesToParse[atv.atvRunesToParsei] = arune
				atv.atvRunesToParsei++
				if atv.atvRunesToParsei == len(atv.atvRunesToParse) {
					captureActiveCode(atv, atv.atvRunesToParse[0:atv.atvRunesToParsei])
					atv.atvRunesToParsei = 0
				}
			} else {
				if strings.TrimSpace(string(arune)) != "" {
					if !atv.foundCode {
						flushPassiveContent(atv, false)
						atv.foundCode = true
					} else {
						flushPassiveContent(atv, false)
					}
					atv.hasCode = true
					if len(atv.atvRunesToParse) == 0 {
						atv.atvRunesToParse = make([]rune, 81920)
					}
					atv.atvRunesToParse[atv.atvRunesToParsei] = arune
					atv.atvRunesToParsei++
					if atv.atvRunesToParsei == len(atv.atvRunesToParse) {
						captureActiveCode(atv, atv.atvRunesToParse[0:atv.atvRunesToParsei])
						atv.atvRunesToParsei = 0
					}
				}
			}
		}
	}
	return
}

func setAtvA(atv *Active, d interface{}) {
	if rdrRune, rdrRuneOk := d.(io.RuneReader); rdrRuneOk {
		atv.rdrRune = rdrRune
	}
	if rdrskr, rdrskrok := d.(io.Seeker); rdrskrok {
		atv.rdskr = rdrskr
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

func NewActive(a ...interface{}) (atv *Active) {
	atv = &Active{maxBufSize: 81920, lck: &sync.RWMutex{}}

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
	if len(atv.runeLabel) > 0 {
		atv.runeLabelI[0] = 0
		atv.runeLabelI[1] = 0
	}
	if len(atv.runePrvR) == 1 {
		atv.runePrvR[0] = rune(0)
	}
	if atv.runesToParsei > 0 {
		atv.runesToParsei = 0
	}
}

func (atv *Active) Close() {
	if len(atv.runeLabel) > 0 {
		atv.runeLabelI = nil
		atv.runeLabel = nil
	}
	if len(atv.runePrvR) == 1 {
		atv.runePrvR = nil
	}
	if atv.runesToParsei > 0 {
		atv.runesToParsei = 0
	}
	if len(atv.runesToParse) > 0 {
		atv.runesToParse = nil
	}
	if atv.rdrRune != nil {
		atv.rdrRune = nil
	}
	if atv.rdskr != nil {
		atv.rdskr = nil
	}
	if atv.printer != nil {
		atv.printer = nil
	}

	//
	if atv.runesToParse != nil {
		atv.runesToParse = nil
	}
	if atv.runeLabel != nil {
		atv.runeLabel = nil
	}
	if atv.runeLabelI != nil {
		atv.runeLabelI = nil
	}
	if atv.runePrvR != nil {
		atv.runePrvR = nil
	}
	//
	if atv.passiveRune != nil {
		atv.passiveRune = nil
	}
	if atv.passiveBuffer != nil {
		for len(atv.passiveBuffer) > 0 {
			atv.passiveBuffer[0] = nil
			atv.passiveBuffer = atv.passiveBuffer[1:]
		}
		atv.passiveBuffer = nil
	}
	//
	if atv.psvRunesToParse != nil {
		atv.psvRunesToParse = nil
	}
	if atv.psvLabel != nil {
		atv.psvLabel = nil
	}
	if atv.psvLabelI != nil {
		atv.psvLabelI = nil
	}
	if atv.psvPrvR != nil {
		atv.psvPrvR = nil
	}
	//
	if atv.atvRunesToParse != nil {
		atv.atvRunesToParse = nil
	}
	if atv.curAtvCde != nil {
		atv.curAtvCde.Close()
		atv.curAtvCde = nil
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
