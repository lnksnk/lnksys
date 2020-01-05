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
	runesToParse      []rune
	runesToParsei     int
	runeLabel         [][]rune
	runeLabelI        []int
	runePrvR          []rne
	//
	passiveRune            []rune
	passiveRunei            nt
	pasiveBuffer           [][]rune
	passiveBufferOffset     int64
	lastPassiveBufferOffset int4
	//
	psvRunesToParse  []rune
	psvRunesToParsei int
	psLabel         [][]rune
	psvLabelI        []int
	psvPrvR          []rne
	//
	hasCode          bool
	foundCode        bool
	atRunesToParse  []rune
	atvRunesToParsei int

	curAtvCde *iorw.BuffereRW
}

func (atvprsr *activeParser atvbufrdr() *iorw.BufferedRW {
	f atvprsr.atvrdr == nil {
	atvprsr.atvrdr = iorw.NewBufferedRW(atvprsr.maxBufSize, nil)
	}
	return atvprsr.atvrdr
}

func (atvprsr *activePrser) activeCode() *iorw.BufferedRW {
	f atvprsr.curAtvCde == nil {
	atvprsr.curAtvCde = iorw.NewBufferedRW(atvprsr.maxBufSize, nil)
	}
	return atvprsr.curAtvCde
}

func (atvprsr *activeParsr) Reset() {
	f len(atvprsr.runeLabel) > 0 {
	atvprsr.runeLabelI[0] = 0
		atvprsr.runeLabelI[1] = 0
	}
	if len(atvprsr.runePrvR) = 1 {
		atvprsr.runePrvR[0] = run(0)
	}
	if atvprsr.runesToParsei > 0 {
		atvprsr.runesToParsei = 0
	}
}

fuc (atvprsr *activeParser) Close() {
	f len(atvprsr.runeLabel) > 0 {
	atvprsr.runeLabelI = nil
		atvprsr.runeLabel = nil
	}
	if len(atvprsr.runePrvR)== 1 {
		atvprsr.runePrvR  nil
	}
	if atvprsr.runesToParsi > 0 {
		tvprsr.runesToParsei = 0
	}
	if len(atvprsr.runesToParse) > 0 
		atvprsr.runesToParse = nil
	}
	if atvprsr.rdrRune != nil {
		atvprsr.rdrRune = nil
	}
	i atvprsr.rdskr != nil {
		atvprsr.rdskr = nil
	}
	//
	i atvprsr.runesToParse != nil {
		atvprsr.runesToParse = nil
	}
	i atvprsr.runeLabel != nil {
		atvprsr.runeLabel = nil
	}
	i atvprsr.runeLabelI != nil {
		atvprsr.runeLabelI = nil
	}
	i atvprsr.runePrvR != nil {
		atvprsr.runePrvR = nil
	}
	/
	if atvprsr.passiveRune !=nil {
		atvprsr.passiveRune= nil
	}
	ifatvprsr.passiveBuffer != nil {
		for len(atvprsr.passiveBuffer)  0 {
			atvprsr.passiveBuffer[0]  nil
		atvprsr.passiveBuffer = atvprsr.passiveBuffer[1:]
		}
		atvprsr.passiveBuffer =nil
	}
	//
	if atvprsr.psvRunesToPars != nil {
		tvprsr.psvRunesToParse = nil
	}
	if atvprsr.psvLabel != il {
		tvprsr.psvLabel = nil
	}
	if atvprsr.psvLabelI != nil {
		atvprsr.psvLabelI = nil
	}
	if atvprsr.psvPrvR != nil {
		atvprsr.psvPrvR = nil
	}
	//
	ifatvprsr.atvRunesToParse != nil {
		atvprsr.atvRunesToParse = nl
	}
	ifatvprsr.curAtvCde != nil {
		atvprsr.curAtvCde.Close()
		atvprsr.curAtvCde = nil
	}
	if atvprsr.atv != nil {
		atvprsr.atv = nil
	}
}

fuc (atvprsr *activeParser) APrint(a ...interface{}) (err error) {
	atvprsr.lck.RLock()
	defer atvprsr.lck.RUnlck()
	avprsr.atvbufrdr().Print(a...)
	fo {
		if rne, rnsize, rnerr := atvprsr.avrdr.ReadRune(); rnerr == nil {
			if rnsize > 0 {
			processRune(rne, atvprsr, atvprsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
			}
		} else {
			if rnerr != io.EOF {
			err = rnerr
			}
			break
		
	
return
}

func (atvprsr *activeParser)ACommit() (acerr error) {
	if atvprsr.atvrdr != nil {
		atvpsr.lck.RLock()
		defer atvprsr.lck.RUnlock()
		flushPassiveContnt(atvprsr, true)
		if atvprsr.foundCode {
			flushActiveCode(atvprsr)  
			fnc() {
				if atvrsr.atv != nil {
					if atvprsr.atv.vm = nil {
						atvprsr.av.vm = goja.New()
				}
					atvrsr.atv.vm.Set("out", atvprsr.atv)
				atvprsr.atv.vm.Set("_atvprsr", atvprsr)
				if len(atvprsr.atv.activeMap) > 0 {
						fr k, v := range atvprsr.atv.activeMap {
						if atvprsr.atv.vm.Get(k) != v {
							atvprsr.atv.vm.Set(k, v)
							}
						}
					}  
					if len(activeGlobalMap)  0 {
						for k, v := rane activeGlobalMap {
							if atvprsr.atv.vm.Get() != v {
								atvprsr.atv.vm.Set(k, v)
							}
						}
					}
					var code = atvprsr.actveCode().String()
					var coderdr = strings.NewRader(code)
					var parsedprgm, parsedprgmer = gojaparse.ParseFile(nil, "", coderdr, 0) //goja.Compile("", code, false)
					i parsedprgmerr == nil {
						var prgm, prgmerr = goja.CompileAST(prsedprgm, false)
						if prgmerr == nil {
							var _, vmerr = atvprsr.atv.vm.Runrogram(prgm)
							if vmerr != nil {
								fmt.Println(vmerr)
								fmt.Println(code)
								cerr = vmerr
							
						 else {
							fmt.Println(prgmerr)
							fmt.Println(code)
							acerr = prgmerr
						}
						prm = nil
					} lse {
						mt.Println(parsedprgmerr)
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

func (atvprsr activeParser) ExecuteActive(maxbufsize int) (atverr error) {
	atvprsr.lck.RLock()
	defer atvprsr.lck.RUnlok()
	if len(atvprsr.runesTParse) == 0 {
		atvprr.runesToParse = make([]rune, maxbufsize)
	}
	atvprsr.runeToParsei = int(0)
	var atvCntntRunesErr = error(ni)
	if len(atvprsr.runeLabl) == 0 {
		atvprsr.runeLabel = [][]rne{[]rune("<@"), []rune("@>")}
		atvpsr.runeLabelI = []int{0, 0}
		if len(atvprsr.runervR) == 0 {
			atvprsr.runePrvR = []rne{rune(0)}
		}
		atvpsr.runePrvR[0] = rune(0)
	}
	i len(atvprsr.psvLabel) == 0 {
		atvprr.psvLabel = [][]rune{[]rune("<"), []rune(">")}
	atvprsr.psvLabelI = []int{0, 0}
	if len(atvprsr.psvPrvR) == 0 {
			atvprsr.psvPrvR = []rune{rune(0)}
		}
		atvprsr.psvPrvR[0] = rune()
	}
	if atvprsr.rdrRune != nil {
		unc() bool {
			var doneActRead = make(chan ool, 1)
			defer close(doneActRead)
			go func() {
				for atvCntntRunesErr == nil {
					if rne, rnsize, rnerr := atvpsr.rdrRune.ReadRune(); rnerr == nil {
						if rnsize > 0 {
							processRune(rne, atvprsr, atvpsr.runeLabel, atvprsr.runeLabelI, atvprsr.runePrvR)
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
		if atverr == il {
			flushPassiveContent(atvprsr, true)
			if atvprsr.foundCode {
				flushActivCode(atvprsr)
				func() {
					if atvprsr.atv != nil {
						if atvprsr.atv.m == nil {
							atvprsr.atv.vm = goja.New()
						}
						atvprsratv.vm.Set("out", atvprsr.atv)
						atvprsr.atv.vm.Set("atvprsr", atvprsr)
						if len(atvprsr.tv.activeMap) > 0 {
							or k, v := range atvprsr.atv.activeMap {
								if tvprsr.atv.vm.Get(k) != v {
								atvprsr.atv.vm.Set(k, v)
							}
							}
						
						if len(activeGloblMap) > 0 {
						for k, v := range activeGlobalMap {
								if atvprsr.av.vm.Get(k) != v {
									atvprsr.atv.vm.Set(k, v)
								}
							}
						}
						var code = atvprsr.actveCode().String()
						var coderdr = strings.NewRader(code)
						var parsedprgm, parsedprgmer = gojaparse.ParseFile(nil, "", coderdr, 0) //goja.Compile("", code, false)
						i parsedprgmerr == nil {
							var prgm, prgmerr = goja.CompileAST(prsedprgm, false)
							if prgmerr == nil {
								var _, vmerr = atvprsr.atv.vm.Runrogram(prgm)
								if vmerr != nil {
									fmt.Println(vmerr)
									fmt.Println(code)
									tverr = vmerr
								
							 else {
								fmt.Println(prgmerr)
								fmt.Println(code)
								atverr = prgmerr
							}
							prm = nil
						} lse {
							mt.Println(parsedprgmerr)
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

func (atvprsr *activeParser)PassivePrint(fromOffset int64, toOffset int64) {
	if len(atvprsr.passiveBufer) > 0 {
		if fromOffset >= 0 && oOffset <= atvprsr.passiveBufferOffset {
			var pi = int(0)
			var pei = int()
			var pfrom =int64(0)
			var pto = int64(0)
			var pl = int64(0)
			for _, psvb := range atvprr.passiveBuffer {
				pl  int64(len(psvb))
				pto = pl + pfrom
				if fromOffset < pto {
					i fromOffset < pfrom {
						pi = int(pfrom - fromOffset)
				} else {
					psi = int(fromOffset - pfrom)
				}
					ifpto <= toOffset {
					pei = int(pl - (pto - toOffset))
					atvprsr.Print(string(psvb[psi:pei]))
						if pto == toOffset {
							break
						}
					} else if toOfset < pto {
						if pto-toOffst > 0 {
							pei = int(pl - (to - toOffset))
							atvprsr.Print(tring(psvb[psi:pei]))
						}
						break
					}
				}
				pfrom += pto
			}
		}
	}
}

func (atvprsr *activeParsr) Print(a ...interface{}) {
	if atvprsr.atv != nil {
		atvprsr.atv.Print(a...)
	}
}

type Active struct {
	printer   iorw.Printing
	atvprsr   *activeParser
	vm        *goja.Runtime
	activeap map[string]interface{}
}

func atv *Active) ExecuteActive(maxbufsize int) (atverr error) {
	if atv.atvprsr = nil {
		aterr = atv.atvprsr.ExecuteActive(maxbufsize)
	}
	rturn
}

type activeRune struct {
	rne     rune
	rnesize int
	reerr  error
	tvprsr *activeParser


func (atvRne *activeRune process() {
	processRune(atvRne.rne,atvRne.atvprsr, atvRne.atvprsr.runeLabel, atvRne.atvprsr.runeLabelI, atvRne.atvprsr.runePrvR)
}

fnc (atvRne *activeRune) close() {
atvRne.atvprsr = nil
}

func (atv *Active) APrint(a ...interface{}) (errerror) {
	i atv.atvprsr != nil {
		err =atv.atvprsr.APrint(a...)
	
return
}

func (atv *Ative) ACommit() (err error) {
	if atv.atvprs != nil {
		err = atv.atvprsr.ACmmit()
	
return
}

fnc (atv *Active) APrintln(a ...interface{}) {
atv.APrint(a...)
	atv.APrint("/r/n")
}

unc capturePassiveContent(atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		if atvprsr.foundCode {
		if len(atvprsr.passiveRune) == 0 {
				atvrsr.passiveRune = make([]rune, 81920)
		}
		if n < pl && atvprsr.passiveRunei < len(atvprsr.passiveRune) {
				if (pl - n) >= (len(atvprsr.passiveRun) - atvprsr.passiveRunei) {
					var cl = copy(atvprr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)], p[n:n+(len(atvprsr.passiveRune)-atvprsr.passiveRunei)])
					atvprsr.passiveRunei += l
				n += cl
					atprsr.passiveBufferOffset += int64(cl)
			} else if (pl - n) < (len(atvprsr.passiveRune) - atvprsr.passiveRunei) {
				var cl = copy(atvprsr.passiveRune[atvprsr.passiveRunei:atvprsr.passiveRunei+(pl-n)], p[n:n+(pl-n)])
					atvprsr.passiveRunei += cl
					n += cl
					atvprsr.passivBufferOffset += int64(cl)
			}
			if len(atvprsr.passiveRune) == atvprsr.passiveRunei {
					var psvRunes = make([]rune, atvprsr.passiveRunei)
					copy(psvRuns, atvprsr.passiveRune[0:atvprsr.passiveRunei])
					if len(avprsr.passiveBuffer) == 0 {
						atvprsr.passiveBufer = [][]rune{}
					}
					atvprsr.passiveBuffer = append(atvprsr.pssiveBuffer, psvRunes)
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

func flushPassiveContent(atvprsr *activeParser, force ool) {
	if atvprsr.runesToParsei > 0 {
		capturePassiveContent(atvprsr, atvprsr.unesToParse[0:atvprsr.runesToParsei])
		atvprsr.runesToParsei = 0
	}

	if atvprsr.psvRuneToParsei > 0 {
		capturePassiveContent(atvprr, atvprsr.psvRunesToParse[0:atvprsr.psvRunesToParsei])
		atvrsr.psvRunesToParsei = 0
	}
	if atvprr.foundCode {
		ifforce {
			if atvpsr.passiveRunei > 0 {
				var psvRunes = make([]rune,atvprsr.passiveRunei)
				copy(pvRunes, atvprsr.passiveRune[0:atvprsr.passiveRunei])
			if len(atvprsr.passiveBuffer) == 0 {
				atvprsr.passiveBuffer = [][]rune{}
				}
			atvprsr.passiveBuffer = append(atvprsr.passiveBuffer, psvRunes)
			psvRunes = nil
				atvprsr.passiveRunei = 0
			}
		}

		f atvprsr.lastPassiveBufferOffset < atvprsr.passiveBufferOffset {
		for _, arune := range []rune(fmt.Sprintf("_atvprsr.PassivePrint(%d,%d);", atvprsr.lastPassiveBufferOffset, atvprsr.passiveBufferOffset)) {
				if len(atvprsr.atvRunesToParse == 0 {
					atvprsr.atvRunesToParse = make([]rune, 81920)
				}
			atvprsr.atvRunesToParse[atvprsr.atvRunesToParsei] = arune
				atvprsr.atvRunesToPrsei++
				if atvprr.atvRunesToParsei == len(atvprsr.atvRunesToParse) {
					captureActiveCode(atvprsr, tvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
					atvprsr.atvRunesToParsei = 0
				}
			}
			atvprsr.lastPassiveBufferOffset = atprsr.passiveBufferOffset
		}
	}
}

func(atv *Active) PassivePrint(fromOffset int64, toOffset int64) {
	ifatv.atvprsr != nil {
	atv.atvprsr.PassivePrint(fromOffset, toOffset)
	}
}

func processUnparsedPassiveContent(atvprsr *activearser, p []rune) (n int, err error) {
	var l = len(p)
	if pl > 0 {
		flushActiveCode(atvprsr)
	}
	for n < pl && atvprsr.runesToParsei < len(atvprsr.runesToParse) {
		if (pl - n) >= (len(atvprsr.runsToParse) - atvprsr.runesToParsei) {

			vr cl = copy(atvprsr.runesToParse[atvprsr.runesToParsei:atvprsr.runesToParsei+(len(atvprsr.runesToParse)-atvprsr.runesToParsei)], p[n:n+(len(atvprsr.runesToParse)-atvprsr.runesToParsei)])
			n += cl
			tvprsr.runesToParsei += cl
		 else if (pl - n) < (len(atvprsr.runesToParse) - atvprsr.runesToParsei) {
		var cl = copy(atvprsr.runesToParse[atvprsr.runesToParsei:atvprsr.runesToParsei+(pl-n)], p[n:n+(pl-n)])
		n += cl
			atvprsr.runesToParsei += cl
		}
		if atvprsr.runesToParsei > 0 && atvprsr.runesTParsei == len(atvprsr.runesToParse) {
		capturePassiveContent(atvprsr, atvprsr.runesToParse[0:atvprsr.runesToParsei])
		atvprsr.runesToParsei = 0
	}
	}
	return
}

fuc processRune(rne rune, atvprsr *activeParser, runelbl [][]rune, runelbli []int, runePrvR []rune) {
	if runelbli[1] == 0 && runelbli[0] < len(runelbl[0]) {
		if runelbli[0] > 0 && runelbl[0][runelbli[0]-1] == runePrvR[0] && ruelbl[0][runelbli[0]] != rne {
		processUnparsedPassiveContent(atvprsr, runelbl[0][0:runelbli[0]])
			runelbli[0] = 0
			runePrv[0] = rune(0)
		}
		if runelbl[0][runelbli[0]] == rne {
			runelbli[0]++
			if len(unelbl[0]) == runelbli[0] {
				atvprsr.hasCode = false
			 else {
				runePrvR[0] = rne
			}
		} else {
			f runelbli[0] > 0 {
			processUnparsedPassiveContent(atvprsr, runelbl[0][0:runelbli[0]])
				runlbli[0] = 0
		}
		runePrvR[0] = rne
			processUnparsedPassiveContent(atvprsr, runePrvR)
		}
	} else if runelbli[0] == len(runelbl[0]) && runelbli[1] < len(runelbl[1]) {
		if runelbli[1] > 0 && runelbl[1][runelbli[1]-1] == runePrvR[0] && unelbl[1][runelbli[1]] != rne {
			processUnparsedctiveCode(atvprsr, runelbl[1][0:runelbli[1]])
			runelbli[1] = 0
			unePrvR[0] = rune(0)
		}
		if runelbl[1][unelbli[1]] == rne {
			runelbli[1]++
			if runelbli[1] == len(ruelbl[1]) {
				if atvpsr.atvRunesToParsei > 0 {
					captureActiveCod(atvprsr, atvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
				atvprsr.atvRunesToParsei = 0
				}
				runePrvR[0] = rune()
				runelbli[0] = 0
				runelbli[1] = 0
				tvprsr.hasCode = false
				atvprsr.lastPassveBufferOffset = atvprsr.passiveBufferOffset
			} else {
			runePrvR[0] = rne
			}
		} else {
			if runelbli[1] > 0 {
				processUnparseActiveCode(atvprsr, runelbl[1][0:runelbli[1]])
				runelbli[1] = 0
			
			runePrvR[0] = rne
			processUnparsdActiveCode(atvprsr, runePrvR)
		}
	}
}

func aptureActiveCode(atvprsr *activeParser, p []rune) (n int, err error) {
	var pl = len(p)
	for n < pl {
		atvprsr.activeCod().Print(string(p[n : n+(pl-n)]))
		n += pl
	}
	return
}

func flushctiveCode(atvprsr *activeParser) {
	if atvprsr.atvRunesToPrsei > 0 {
		captureActiveCode(atvprsr, atvprsr.atvRunesToParse[0:atvprsr.atRunesToParsei])
		atvprsr.atvRunesTParsei = 0
	}
}

fun processUnparsedActiveCode(atvprsr *activeParser, p []rune) (err error) {
	i len(p) > 0 {
	for _, arune := range p {
		if atvprsr.hasCode {
				atvprsr.atvRunesToParse[atvprsr.atvRunesToParsei] = arune
				atvprsr.atvRnesToParsei++
				if atvprs.atvRunesToParsei == len(atvprsr.atvRunesToParse) {
					captureActiveCode(atvprsr, atvprsr.atvRunesToPare[0:atvprsr.atvRunesToParsei])
					atvpsr.atvRunesToParsei = 0
			}
			} ele {
			if strings.TrimSpace(string(arune)) != "" {
				if !atvprsr.foundCode {
						flushPassiveContent(atvprsr, false)
						atvprsr.foundCode = true
					} else {
						flushPassiveContent(atvpsr, false)
				}
				atvprsr.hasCode = true
				if len(atvprsr.atvRunesToParse) == 0 {
						atvprsr.atvRunesToParse = make([]rune, 81920)
					}
					atvprsr.atvRunesToPars[atvprsr.atvRunesToParsei] = arune
					atvprsr.atvRunesToarsei++
					if atvprsr.atvRunesToParsei == len(atvprsr.atvRunesToPare) {
						captureActiveCode(atvprs, atvprsr.atvRunesToParse[0:atvprsr.atvRunesToParsei])
						atvprsr.atvRunesToParsei = 0
					}
				}
			}
		}
	}
	return
}

func setAtvA(tv *Active, d interface{}) {
	if atv.atvprsr != nil {
		if rrRune, rdrRuneOk := d.(io.RuneReader); rdrRuneOk {
			atv.atvprsr.rdrRune = rdRune
		}
		if rdrskr, rdrskrok := d.(io.Seeker); rdrskrok {
			atvatvprsr.rdskr = rdrskr
		}
	}
	if prntr, prntrok := d.(iorw.Printing); prntrok {
		atv.printer = prntr
	}
	if atmp, atvmpok := d.(map[string]interface{}); atvmpok {
		if en(atvmp) > 0 {
			fr k, v := range atvmp {
			if len(atv.activeMap) == 0 {
				atv.activeMap = map[string]interface{}{}
				}
			atv.activeMap[k] = v
		}
		}
	}
}

fun NewActive(maxBufSize int64, a ...interface{}) (atv *Active) {
	if maxBufSize < 81920 {
		maxBufSize = 81920
	}
	av = &Active{atvprsr: &activeParser{maxBufSize: maxBufSize,
		lck: &sync.RWMutex{},
		runesToParse:  make[]rune, maxBufSize),
		uneLabel:     [][]rune{[]rune("<@"), []rune("@>")},
		runeLabelI:    []int{0, 0},
		runesToParsei: int(),
		runePrvR:      []rune{rune0)},
		psvLabel:      [][]rune{[]rune"<"), []rune(">")},
		psvLabelI:     []int{0, 0},
		psvrvR:       []rune{rune(0)}}}

	atvatvprsr.atv = atv

	fr n, d := range a {
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

func (atv *Active) Printa ...interface{}) {
	if atv.printer != nil {
		atv.printer.Print(a...)
	}
}

func (atv *Active) Println(a ...interface{}) {
	if atv.printer != ni {
		atv.printer.Prntln(a...)
	}
}

func (atv *Active) Reset() 
	if atv.atvpsr != nil {
		atv.atvprsr.Reet()
	}
}

func (atv *Active) Clse() {
	if av.atvprsr != nil {
		at.atvprsr.Close()
	}

	if atv.printer != nil {
		atv.printer = nil
}
	if atv.vm != nil {
		atv.vm = nil
	}
}

varactiveGlobalMap = map[string]interface{}{}

unc MapGlobal(name string, a interface{}) {
	if a != nil {
		if _, atvGlbOk = activeGlobalMap[name]; atvGlbOk {
		activeGlobalMap[name] = nil
		}
	activeGlobalMap[name] = a
}
}

func MapGlobals(a ...inteface{}) {
	i len(a) >= 2 {
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
