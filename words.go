package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilwo/wordscvc/cvc"
	"github.com/jessevdk/go-flags"
)

type flagOpts struct {
	// flag able vars
	MaxGroups                   int    `short:"G" description:"1  number of result groups to generate" default:"20"`
	MaxSets                     int    `short:"S" description:"2  number of sets per group" default:"15"`
	MaxWords                    int    `short:"W" description:"3  number of words per set" default:"10"`
	FreqCutoff                  int    `short:"f" description:"4  frequency cutoff threshold for words, lower is more common" default:"25"`
	FreqWordsPerLineAboveCutoff int    `short:"a" description:"5  how many words to be above cutoff threshold per line" default:"3"`
	VowelLimit                  int    `long:"vowel" description:"6  how many time each vowel repeat per set" hidden:"1"`// 2

	InConsonantFile             string `short:"C" description:"7  input file name for consonants to use" optional:"1" default:"consonants.txt"`
	InVowelFile                 string `short:"V" description:"8  input file name for vowels to use" optional:"1" default:"vowels.txt"`
	InWordsFile                 string `short:"i" description:"9  input file name for words list to use for creating the lines groups results" optional:"1" default:"words_list.txt"`

	FilterFile                  string `short:"F" description:"10 input file name for filtered words"`
	OutResultFile               string `short:"o" description:"11 output file for generated results" default:"words_result.txt" default-mask:"-"`

	TimeToRun                   int    `short:"t" description:"12 how much time to run (in seconds)" default:"30"`

	CpuProfile                  string `short:"c" description:"13 enable cpu profiling and save to file"`
	MemProfile                  string `short:"m" description:"14 enable memory profiling and save to file"`

	DebugEnabled                bool   `short:"d" description:"15 enable debugging information"`
	Verbose                     [2]bool `short:"v" description:"16 show verbose information" choice:"2"`
}

func (fo flagOpts) String() string {
	return fmt.Sprintf("options setting:\n"+
		"\tmax groups: '%v'\n"+
		"\tmax sets: '%v'\n"+
		"\tmax words: '%v'\n"+
		"\tfrequency cutoff: '%v'\n"+
		"\tabove frequency words per set: '%v'\n"+
		//"\tvowels limit: '%v'\n"+
		"\n"+
		"\tconsonant file : '%v'\n"+
		"\tvowels file: '%v'\n"+
		"\twords file: '%v'\n"+
		"\n"+
		"\tfilter file: '%v'\n"+
		"\tresult output file: '%v'\n"+
		"\n"+
		"\ttime to run: '%v'\n"+
		"\n"+
		"\tcpu profile file: '%v'\n"+
		"\tmemory profile file: '%v'\n"+
		"\tdebug enabled: '%v'\n"+
		"\tverbose: '%v'\n",
		fo.MaxGroups,
		fo.MaxSets,
		fo.MaxWords,
		fo.FreqCutoff,
		fo.FreqWordsPerLineAboveCutoff,
		//fo.VowelLimit,
		fo.InConsonantFile,
		fo.InVowelFile,
		fo.InWordsFile,
		fo.FilterFile,
		fo.OutResultFile,
		fo.TimeToRun,
		fo.CpuProfile,
		fo.MemProfile,
		fo.DebugEnabled,
		fo.Verbose,
	)
}

type varOpts struct {
	flagOpts

	// internal vars
	countGroups    int
	finishSignal   bool
	maxWorkers     int
	currentWorkers int
}

var GenVarOpts varOpts

var consonants, vowels map[string]int

var waitForWorkers = make(chan bool)
var collectingDone = make(chan struct{})
var msgs = make(chan string, 100)
var startedWorkers = make(chan struct{}, 100)
var stoppedWorkers = make(chan struct{}, 100)
var maxSize int = 0

type findArg struct {
	group *cvc.GroupSet
	wordmap *cvc.WordMap
}

func findGroups(arg findArg) {
	defer func() {
		if fail := recover(); fail != nil {
			// fmt.Printf("recovered from %s\n", fail)
		}
		stoppedWorkers <- struct{}{}
		arg.group = nil
		arg.wordmap = nil
		// runtime.GC()
		return
	}()

	startedWorkers <- struct{}{}
	zmap := *arg.wordmap.GetCm()
	if !arg.group.Checkifavailable(arg.wordmap) {
		return
	}
	if float64(arg.group.CurrentSize())/float64(arg.group.MaxSize()) > float64(0.9) {
		s := fmt.Sprintf("status: reached depth %d of %d\n",
			int(arg.group.CurrentSize()), int(arg.group.MaxSize()))
		msgs <- s
	}
	if arg.group.CurrentSize() > maxSize {
		msgs <- "depth: " + strconv.Itoa(arg.group.CurrentSize())
	}

	for k := range zmap {

		if GenVarOpts.finishSignal {
			// fmt.Println("finishSignal issued, exiting")
			break
		}
		if GenVarOpts.countGroups >= GenVarOpts.MaxGroups {
			// fmt.Printf("groups count %d reached max groups %d", GenVarOpts.countGroups, GenVarOpts.MaxGroups)
			break
		}
		if added, full := arg.group.AddWord(k); full == true {
			msg := fmt.Sprintf("group completed\n%s\n", arg.group.StringWithFreq())
			if GenVarOpts.DebugEnabled {
				msg += arg.group.DumpGroup() + "\n"
			}
			// fmt.Println(msg)
			msgs <- msg
			break
		} else if added {
			arg.wordmap.DelWord(k)
			go findGroups(findArg{arg.group.CopyGroupSet(), arg.wordmap.CopyWordMap()})
		}
	}
}

func main() {

	a, err := flags.NewParser(&GenVarOpts, flags.Default).Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			switch e.Type {
			case flags.ErrHelp:
				//fmt.Println("help requested, existing")
				os.Exit(0)
			case flags.ErrInvalidChoice:
				fmt.Println("-v or -vv")
				os.Exit(1)
			default:
				fmt.Printf("error parsing opts: %v\n", e.Type)
			}
		}
	}
	fmt.Printf("opts : %v\na %v\n", GenVarOpts, a)

	if GenVarOpts.MemProfile != "" {
		f, err := os.Create(GenVarOpts.MemProfile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("enable memprofiling, write to '%v'\n", GenVarOpts.MemProfile)
		defer func() {
			pprof.WriteHeapProfile(f)
			f.Close()
			return
		}()
	}

	if GenVarOpts.CpuProfile != "" {
		f, err := os.Create(GenVarOpts.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("enable cpuprofiling, write to '%v'\n", GenVarOpts.CpuProfile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	consonants = getMap(GenVarOpts.InConsonantFile)
	vowels = getMap(GenVarOpts.InVowelFile)
	fmt.Printf("consonants: %d\n%s\n", len(consonants), getOrderedMapString(consonants))
	fmt.Printf("vowels: %d\n%s\n", len(vowels), getOrderedMapString(vowels))

	wmap := getWordsMap(GenVarOpts.InWordsFile)
	fmt.Printf("map size: %d\ncontent:\n%s\n", wmap.Size(), wmap)

	// set the base group according to the required settings
	baseGroup := cvc.NewGroupSetLimitFreq(
		GenVarOpts.MaxSets,
		GenVarOpts.MaxWords,
		GenVarOpts.FreqCutoff,
		GenVarOpts.FreqWordsPerLineAboveCutoff)

	// start time measuring
	t0 := time.Now()

	// wait for all goroutines to finish
	go func() {
		count := 0
		i := 0
		for {
			// fmt.Println("going to wait")
			select {
			case <-startedWorkers:
				// println("startedWorkers")
				count++
				if count > GenVarOpts.maxWorkers {
					GenVarOpts.maxWorkers = count
				}
				GenVarOpts.currentWorkers = count
				i = 0
			case <-stoppedWorkers:
				// println("stoppedWorkers")
				count--
				GenVarOpts.currentWorkers = count
				i = 0
			case <-time.After(1 * time.Second):
				if count > 0 {
					fmt.Printf("there are still %d active go routines\n", count)
					// fmt.Printf("count = %d and i = %d", count, i)
				} else {
					fmt.Printf("count = 0 and i = %d\n", i)
					i++
					if i == 3 {
						waitForWorkers <- true
					}
				}
			}
		}
	}()

	// msg collector
	go func() {
		for {
			select {
			case s := <-msgs:
				if strings.HasPrefix(s, "depth: ") {
					size, _ := strconv.Atoi(s[len("depth: "):])
					if size > maxSize {
						maxSize = size
						fmt.Printf("max depth : %d\n", maxSize)
					}
				} else if strings.HasPrefix(s, "status:") {
					fmt.Printf("%s", s)
				} else {
					GenVarOpts.countGroups++
					fmt.Printf("%d\n%s", GenVarOpts.countGroups, s)
					if GenVarOpts.countGroups == GenVarOpts.MaxGroups {
						close(msgs)
						close(collectingDone)
						return
					}
				}
			default:
				time.Sleep(1 * time.Second)
				fmt.Printf("%s passed\n", time.Now().Sub(t0))
				if GenVarOpts.finishSignal {
					// fmt.Println("finishSignal issued, exiting")
					close(msgs)
					return
				}
				if GenVarOpts.DebugEnabled {
					fmt.Printf("current workers %d, max workers %d\n",
						GenVarOpts.currentWorkers, GenVarOpts.maxWorkers)
				}
			}
		}
	}()

	go findGroups(findArg{baseGroup, wmap})

	dur := time.Duration(GenVarOpts.TimeToRun)
	select {
	case <-collectingDone:
		fmt.Printf("required results collected after %s\n", time.Now().Sub(t0))
	case <-time.After(dur * time.Second):
		fmt.Printf("stopped after %s\n", time.Now().Sub(t0))
	}
	GenVarOpts.finishSignal = true

	fmt.Printf("waiting for waitForWorkers, %d workers\n", GenVarOpts.currentWorkers)
	<-waitForWorkers
	fmt.Printf("exiting... after %s\n", time.Now().Sub(t0))

}

func getOrderedMapString(m map[string]int) string {
	out := ""
	var sortedkeys []string
	for k := range m {
		sortedkeys = append(sortedkeys, k)
	}
	sort.Slice(sortedkeys,
		func(i, j int) bool {
			return strings.Compare(sortedkeys[i], sortedkeys[j]) == -1
		})

	for _, k := range sortedkeys {
		out += fmt.Sprintf("%s, ", k)
	}
	return out[:len(out)-2]
	// return out
}

func getMap(mapfile string) map[string]int {
	var ret = make(map[string]int)
	for _, wf := range getWordsFromFile(mapfile) {
		ret[wf.word] = wf.number
	}
	return ret
}

func getWordsMap(fname string) *cvc.WordMap {
	wmap := cvc.NewWordMap()

	for _, wf := range getWordsFromFile(fname) {
		var cvcw *cvc.Word
		wfV := string(wf.word[1])
		if _, ok := vowels[wfV]; ok {
			cvcw = cvc.NewWord(
				string(wf.word[0]),
				string(wf.word[1]),
				string(wf.word[2:]),
				wf.number)
		} else {
			cvcw = cvc.NewWord(
				string(wf.word[0:2]),
				string(wf.word[2]),
				string(wf.word[3:]),
				wf.number)
		}

		if wf.word != cvcw.String() {
			panic("loaded word: " + wf.word + " and built word: " +
				cvcw.String() + " are NOT the same")
		}

		wmap.AddWord(cvcw)
	}
	return wmap
}

func checkErr(e error) {
	if e != nil {
		panic(e)
	}
}

// WF - word number bundle
type WF struct {
	word   string
	number int
}

func getLinesFromFile(fname string) []string {
	data, err := ioutil.ReadFile(fname)
	checkErr(err)

	data = bytes.TrimRight(data, "\n")
	lines := strings.Split(string(data), "\n")
	return lines
}

func getWordsFromFile(fname string) []WF {
	resList := []WF{}

	for _, line := range getLinesFromFile(fname) {
		tmp := strings.Split(line, " ")
		w := strings.TrimRight(tmp[0], ":")
		f, _ := strconv.Atoi(tmp[1])
		resList = append(resList, WF{w, f})
	}
	return resList
}

// var consonants = map[string]int{
// 	"B": 0, "G": 0, "D": 0, "V": 0, "Z": 0, "X": 0, "T": 0,
// 	"J": 0, "K": 0, "L": 0, "M": 0, "N": 0, "S": 0, "P": 0, "F": 0, "TZ": 0,
// 	"R": 0, "SH": 0, "Q": 0, "W": 0, "H": 0}
// var vowels = map[string]int{
// 	"A": 0, "E": 0, "I": 0, "O": 0, "U": 0}
