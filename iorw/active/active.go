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
	runesToParseQueue chan rune
	commitParsedQueue chan bool
	closing           chan bool
	runesToParse      []rune
	runesToParsei     int
	runeLabel         [][]rune
	runeLabelI        []int
	runePrvR          []rune
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
}

func (atvprsr *activeParser) atvbufrdr() *iorw.BufferedRW {
	if atvprsr.atvrdr == nil {
		atvprsr.atvrdr = iorw.NewBufferedRW(atvprsr.maxBufSize, nil)
	}
	return atvprsr.atvrdr
}

func (atvprsr *activeParser) activeCode() *iorw.BufferedRW {
	if atvprsr.curAtvCde == nil {
		atvprsr.curAtvCde = iorw.NewBufferedRW(atvprsr.maxBufSize, nil)
	}
	return atvprsr.curAtvCde
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
	if atvprsr.closing != nil {
		atvprsr.closing <- true
		<-atvprsr.closing
		close(atvprsr.closing)
		atvprsr.closing = nil
	}
	if atvprsr.runesToParseQueue != nil {
		close(atvprsr.runesToParseQueue)
		atvprsr.runesToParseQueue = nil
	}
	if atvprsr.commitParsedQueue != nil {
		close(atvprsr.commitParsedQueue)
		atvprsr.commitParsedQueue = nil
	}
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
	if atvprsr.atvRunesToParse != nil {
		atvprsr.atvRunesToParse = nil
	}
	if atvprsr.curAtvCde != nil {
		atvprsr.curAtvCde.Close()
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
				//atvprsr.runesToParseQueue <- rne
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

func (atvprsr *activeParser) ACommit() (acerr error) {
	if len(atvprsr.runesToParse) == 0 {
		atvprsr.runesToParse = make([]rune, atvprsr.maxBufSize)
		atvprsr.runesToParsei = int(0)
	}

	if atvprsr.atvrdr != nil {
		//atvprsr.commitParsedQueue <- true
		//if <-atvprsr.commitParsedQueue {
		flushPassiveContent(atvprsr, true)
		if atvprsr.foundCode {
			flushActiveCode(atvprsr)
			func() {
				if atvprsr.atv != nil {
					if atvprsr.atv.vm == nil {
						atvprsr.atv.vm = goja.New()
					}
					atvprsr.atv.vm.Set("out", atvprsr.atv)
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
					var code = atvprsr.activeCode().String()
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
		//}
	}
	return
}

func (atvprsr *activeParser) ExecuteActive(maxbufsize int) (atverr error) {
	atvprsr.lck.RLock()
	defer atvprsr.lck.RUnlock()
	if len(atvprsr.runesToParse) == 0 {
		atvprsr.runesToParse = make([]rune, maxbufsize)
	}
	atvprsr.runesToParsei = int(0)
	var atvCntntRunesErr = error(nil)
	if len(atvprsr.runeLabel) == 0 {
		atvprsr.runeLabel = [][]rune{[]rune("<@"), []rune("@>")}
		atvprsr.runeLabelI = []int{0, 0}
		if len(atvprsr.runePrvR) == 0 {
			atvprsr.runePrvR = []rune{rune(0)}
		}
		atvprsr.runePrvR[0] = rune(0)
	}
	if len(atvprsr.psvLabel) == 0 {
		atvprsr.psvLabel = [][]rune{[]rune("<"), []rune(">")}
		atvprsr.psvLabelI = []int{0, 0}
		if len(atvprsr.psvPrvR) == 0 {
			atvprsr.psvPrvR = []rune{rune(0)}
		}
		atvprsr.psvPrvR[0] = rune(0)
	}
	if atvprsr.rdrRune != nil {
		func() bool {
			var doneActRead = make(chan bool, 1)
			defer close(doneActRead)
			go func() {
				for atvCntntRunesErr == nil {
					if rne, rnsize, rnerr := atvprsr.rdrRune.ReadRune(); rnerr == nil {
						if rnsize > 0 {
							processRune(rne, atvprsr, atvprsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
						}
					} else {
						if rnerr != io.EOF {
							atverr = rnerr
						}
						break
					}
				}
				doneActRead <- true
			}()
			return <-doneActRead
		}()
		if atverr == nil {
			flushPassiveContent(atvprsr, true)
			if atvprsr.foundCode {
				flushActiveCode(atvprsr)
				func() {
					if atvprsr.atv != nil {
						if atvprsr.atv.vm == nil {
							atvprsr.atv.vm = goja.New()
						}
						atvprsr.atv.vm.Set("out", atvprsr.atv)
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
						var code = atvprsr.activeCode().String()
						var coderdr = strings.NewReader(code)
						var parsedprgm, parsedprgmerr = gojaparse.ParseFile(nil, "", coderdr, 0) //goja.Compile("", code, false)
						if parsedprgmerr == nil {
							var prgm, prgmerr = goja.CompileAST(parsedprgm, false)
							if prgmerr == nil {
								var _, vmerr = atvprsr.atv.vm.RunProgram(prgm)
								if vmerr != nil {
									fmt.Println(vmerr)
									fmt.Println(code)
									atverr = vmerr
								}
							} else {
								fmt.Println(prgmerr)
								fmt.Println(code)
								atverr = prgmerr
							}
							prgm = nil
						} else {
							fmt.Println(parsedprgmerr)
							fmt.Println(code)
							atverr = parsedprgmerr
						}
						parsedprgm = nil
						atvprsr.atv.vm = nil
					}
				}()
			}
		}
	}
	return
}

func (atvprsr *activeParser) PassivePrint(fromOffset int64, toOffset int64) {
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
						atvprsr.Print(string(psvb[psi:pei]))
						if pto == toOffset {
							break
						}
					} else if toOffset < pto {
						if pto-toOffset > 0 {
							pei = int(pl - (pto - toOffset))
							atvprsr.Print(string(psvb[psi:pei]))
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

func (atv *Active) ExecuteActive(maxbufsize int) (atverr error) {
	if atv.atvprsr != nil {
		atverr = atv.atvprsr.ExecuteActive(maxbufsize)
	}
	return
}

type activeRune struct {
	rne     rune
	rnesize int
	rneerr  error
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
						atvprsr.passiveBuffer = [][]rune{}
					}
					atvprsr.passiveBuffer = append(atvprsr.passiveBuffer, psvRunes)
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

func flushPassiveContent(atvprsr *activeParser, force bool) {
	if atvprsr.runesToParsei > 0 {
		capturePassiveContent(atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
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
					atvprsr.passiveBuffer = [][]rune{}
				}
				atvprsr.passiveBuffer = append(atvprsr.passiveBuffer, psvRunes)
				psvRunes = nil
				atvprsr.passiveRunei = 0
			}
		}

		if atvprsr.lastPassiveBufferOffset < atvprsr.passiveBufferOffset {
			for _, arune := range []rune(fmt.Sprintf("_atvprsr.PassivePrint(%d,%d);", atvprsr.lastPassiveBufferOffset, atvprsr.passiveBufferOffset)) {
				if len(atvprsr.atvRunesToParse) == 0 {
					atvprsr.atvRunesToParse = make([]rune, 81920)
				}
				atvprsr.atvRunesToParse[atvprsr.atvRunesToParsei] = arune
				atvprsr.atvRunesToParsei++
				if atvprsr.atvRunesToParsei == len(atvprsr.atvRunesToParse) {
					captureActiveCode(atvprsr, atvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
					atvprsr.atvRunesToParsei = 0
				}
			}
			atvprsr.lastPassiveBufferOffset = atvprsr.passiveBufferOffset
		}
	}
}

func (atv *Active) PassivePrint(fromOffset int64, toOffset int64) {
	if atv.atvprsr != nil {
		atv.atvprsr.PassivePrint(fromOffset, toOffset)
	}
}

func processUnparsedPassiveContent(atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	if pl > 0 {
		flushActiveCode(atvprsr)
	}
	for n < pl && atvprsr.runesToParsei < len(atvprsr.runesToParse) {
		if (pl - n) >= (len(atvprsr.runesToParse) - atvprsr.runesToParsei) {

			var cl = copy(atvprsr.runesToParse[atvprsr.runesToParsei:atvprsr.runesToParsei+(len(atvprsr.runesToParse)-atvprsr.runesToParsei)], p[n:n+(len(atvprsr.runesToParse)-atvprsr.runesToParsei)])
			n += cl
			atvprsr.runesToParsei += cl
		} else if (pl - n) < (len(atvprsr.runesToParse) - atvprsr.runesToParsei) {
			var cl = copy(atvprsr.runesToParse[atvprsr.runesToParsei:atvprsr.runesToParsei+(pl-n)], p[n:n+(pl-n)])
			n += cl
			atvprsr.runesToParsei += cl
		}
		if atvprsr.runesToParsei > 0 && atvprsr.runesToParsei == len(atvprsr.runesToParse) {
			capturePassiveContent(atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
			atvprsr.runesToParsei = 0
		}
	}
	return
}

func processRune(rne rune, atvprsr *activeParser, runelbl [][]rune, runelbli []int, runePrvR []rune) {
	if runelbli[1] == 0 && runelbli[0] < len(runelbl[0]) {
		if runelbli[0] > 0 && runelbl[0][runelbli[0]-1] == runePrvR[0] && runelbl[0][runelbli[0]] != rne {
			processUnparsedPassiveContent(atvprsr, runelbl[0][0:runelbli[0]])
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
				processUnparsedPassiveContent(atvprsr, runelbl[0][0:runelbli[0]])
				runelbli[0] = 0
			}
			runePrvR[0] = rne
			processUnparsedPassiveContent(atvprsr, runePrvR)
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
				if atvprsr.atvRunesToParsei > 0 {
					captureActiveCode(atvprsr, atvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
					atvprsr.atvRunesToParsei = 0
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

func captureActiveCode(atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		atvprsr.activeCode().Print(string(p[n : n+(pl-n)]))
		n += pl
	}
	return
}

func flushActiveCode(atvprsr *activeParser) {
	if atvprsr.atvRunesToParsei > 0 {
		captureActiveCode(atvprsr, atvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
		atvprsr.atvRunesToParsei = 0
	}
}

func processUnparsedActiveCode(atvprsr *activeParser, p []rune) (err error) {
	if len(p) > 0 {
		for _, arune := range p {
			if atvprsr.hasCode {
				atvprsr.atvRunesToParse[atvprsr.atvRunesToParsei] = arune
				atvprsr.atvRunesToParsei++
				if atvprsr.atvRunesToParsei == len(atvprsr.atvRunesToParse) {
					captureActiveCode(atvprsr, atvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
					atvprsr.atvRunesToParsei = 0
				}
			} else {
				if strings.TrimSpace(string(arune)) != "" {
					if !atvprsr.foundCode {
						flushPassiveContent(atvprsr, false)
						atvprsr.foundCode = true
					} else {
						flushPassiveContent(atvprsr, false)
					}
					atvprsr.hasCode = true
					if len(atvprsr.atvRunesToParse) == 0 {
						atvprsr.atvRunesToParse = make([]rune, 81920)
					}
					atvprsr.atvRunesToParse[atvprsr.atvRunesToParsei] = arune
					atvprsr.atvRunesToParsei++
					if atvprsr.atvRunesToParsei == len(atvprsr.atvRunesToParse) {
						captureActiveCode(atvprsr, atvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
						atvprsr.atvRunesToParsei = 0
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
	atv = &Active{atvprsr: &activeParser{closing: make(chan bool, 1),
		runesToParseQueue: make(chan rune, 1),
		commitParsedQueue: make(chan bool, 1),
		maxBufSize:        maxBufSize, lck: &sync.RWMutex{},
		runesToParse:  make([]rune, maxBufSize),
		runeLabel:     [][]rune{[]rune("<@"), []rune("@>")},
		runeLabelI:    []int{0, 0},
		runesToParsei: int(0),
		runePrvR:      []rune{rune(0)},
		psvLabel:      [][]rune{[]rune("<"), []rune(">")},
		psvLabelI:     []int{0, 0},
		psvPrvR:       []rune{rune(0)}}}

	go func(prsr *activeParser, prsreRuneQueue chan rune, commitNow chan bool, closeNow chan bool) {
		var isActive = true
		for isActive {
			select {
			case prsrrne := <-prsreRuneQueue:
				processRune(prsrrne, prsr, prsr.runeLabel, prsr.runeLabelI, prsr.runePrvR)
			case cmt := <-commitNow:
				if cmt {
					<-commitNow
				}
			case dne := <-closeNow:
				if dne {
					isActive = false
				}
			}
		}
		closeNow <- true
	}(atv.atvprsr, atv.atvprsr.runesToParseQueue, atv.atvprsr.commitParsedQueue, atv.atvprsr.closing)
	atv.atvprsr.atv = atv
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
