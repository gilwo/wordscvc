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
	"github.com/gilwo/workqueue/pool"
	"github.com/jessevdk/go-flags"
)

type flagOpts struct {
	// flag able vars
	MaxGroups                   int     `short:"G" description:"1  number of result groups to generate" default:"20"`
	MaxSets                     int     `short:"S" description:"2  number of sets per group" default:"15"`
	MaxWords                    int     `short:"W" description:"3  number of words per set" default:"10"`
	FreqCutoff                  int     `short:"f" description:"4  frequency cutoff threshold for words, lower is more common" default:"25"`
	FreqWordsPerLineAboveCutoff int     `short:"a" description:"5  how many words to be above cutoff threshold per line" default:"3"`
	VowelLimit                  int     `long:"vowel" description:"6  how many time each vowel repeat per set" hidden:"1"`// 2

	InConsonantFile             string  `short:"C" description:"7  input file name for consonants to use" optional:"1" default:"consonants.txt"`
	InVowelFile                 string  `short:"V" description:"8  input file name for vowels to use" optional:"1" default:"vowels.txt"`
	InWordsFile                 string  `short:"i" description:"9  input file name for words list to use for creating the lines groups results" optional:"1" default:"words_list.txt"`

	FilterFile                  string  `short:"F" description:"10 input file name for filtered words"`
	OutResultFile               string  `short:"o" description:"11 output file for generated results" default:"words_result.txt" default-mask:"-"`

	TimeToRun                   int     `short:"t" description:"12 how much time to run (in seconds)" default:"30"`

	CpuProfile                  string  `short:"c" description:"13 enable cpu profiling and save to file"`
	MemProfile                  string  `short:"m" description:"14 enable memory profiling and save to file"`

	DebugEnabled                bool    `short:"d" description:"15 enable debugging information"`
	Verbose                     []bool  `short:"v" description:"16 show verbose information"`

	UsePool                     bool    `short:"p" description:"17 enable using the worker pool logic" hidden:"1"`
	UseJobDispose               bool    `short:"D" description:"18 enable using the worker job dispose logic" hidden:"1"`
	Workers                     uint    `short:"w" description:"19 how many workers to use" default:"30" hidden:"1"`

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
		"\tworkers: '%v'\n"+
		"\tuse pool: '%v'\n"+
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
		fo.Workers,
		fo.UsePool,
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
var pool *workerpool.WPool

var consonants, vowels map[string]int

var waitForWorkers = make(chan bool)
var collectingDone = make(chan struct{})
var msgs = make(chan string, 100)
var startedWorkers = make(chan struct{}, 100)
var stoppedWorkers = make(chan struct{}, 100)
var maxSize int = 0
var disposeChan = make (chan *workerpool.WorkerJob, 1000)
var disposeDone = make (chan bool)

type findArg struct {
	group *cvc.GroupSet
	wordmap *cvc.WordMap
}

func findGroups(iarg interface{}, job *workerpool.WorkerJob, stop workerpool.CheckStop) (none interface{}) {
	arg, _ := iarg.(findArg)

	defer func() {
		if fail := recover(); fail != nil {
			verbose("recovered from %s\n", fail)
		}
		stoppedWorkers <- struct{}{}
		arg.group = nil
		arg.wordmap = nil
		if GenVarOpts.UsePool && GenVarOpts.UseJobDispose {
			if !GenVarOpts.finishSignal {
				disposeChan <- job
			}
		}

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

Loop:
	for k := range zmap {

		if GenVarOpts.finishSignal {
			 info("finishSignal issued, exiting\n")
			break Loop
		}
		if GenVarOpts.countGroups >= GenVarOpts.MaxGroups {
			info("groups count %d reached max groups %d", GenVarOpts.countGroups, GenVarOpts.MaxGroups)
			break Loop
		}
		if added, full := arg.group.AddWord(k); full == true {
			msg := fmt.Sprintf("group completed\n%s\n", arg.group.StringWithFreq())
			if GenVarOpts.DebugEnabled {
				msg += arg.group.DumpGroup() + "\n"
			}
			msgs <- msg
			break Loop
		} else if added {
			arg.wordmap.DelWord(k)
			if !GenVarOpts.UsePool {
				go findGroups(
					findArg{
						arg.group.CopyGroupSet(),
						arg.wordmap.CopyWordMap(),
					}, nil, nil)
			} else {
				_, err := pool.NewJobQueue(findGroups,
					findArg{
						arg.group.CopyGroupSet(),
						arg.wordmap.CopyWordMap(),
					})
				if err !=nil {
					info("error queuing job %v\n", err)
				}
				trace("%v\n", pool.PoolStats())
			}
		}
	}
	return
}

func main() {

	var out string

	a, err := flags.NewParser(&GenVarOpts, flags.Default).Parse()
	if err != nil {
		if e, ok := err.(*flags.Error); ok {
			switch e.Type {
			case flags.ErrHelp:
				os.Exit(0)
			default:
				fmt.Printf("error parsing opts: %v\n", e.Type)
				os.Exit(1)
			}
		}
	}
	info("opts:\n%v\na:\n%v\n", GenVarOpts, a)

	if GenVarOpts.UsePool {
		if GenVarOpts.DebugEnabled {
			workerpool.WorkerPoolSetLogLevel(workerpool.DebugLevel)
		}
		if pool, err = workerpool.NewWPool(GenVarOpts.Workers); err != nil {
			fmt.Printf("failed to create pool %v\n", err)
			os.Exit(1)
		}

		if _, err = pool.StartDispatcher(); err != nil {
			fmt.Printf("failed to start pool dispatcher %v\n", err)
			os.Exit(1)
		}
	}

	if GenVarOpts.MemProfile != "" {
		f, err := os.Create(GenVarOpts.MemProfile)
		if err != nil {
			log.Fatal(err)
		}
		info("enable memprofiling, write to '%v'\n", GenVarOpts.MemProfile)
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
		info("enable cpuprofiling, write to '%v'\n", GenVarOpts.CpuProfile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	consonants = getMap(GenVarOpts.InConsonantFile)
	vowels = getMap(GenVarOpts.InVowelFile)
	verbose("consonants: %d\n%s\n", len(consonants), getOrderedMapString(consonants))
	verbose("vowels: %d\n%s\n", len(vowels), getOrderedMapString(vowels))

	wmap := getWordsMap(GenVarOpts.InWordsFile)
	verbose("map size: %d\ncontent:\n%s\n", wmap.Size(), wmap)

	// set the base group according to the required settings
	baseGroup := cvc.NewGroupSetLimitFreq(
		GenVarOpts.MaxSets,
		GenVarOpts.MaxWords,
		GenVarOpts.FreqCutoff,
		GenVarOpts.FreqWordsPerLineAboveCutoff)

	// start time measuring
	t0 := time.Now()

	if GenVarOpts.UsePool && GenVarOpts.UseJobDispose {
		// job disposer
		go func() {
			for j := range disposeChan {
				if j.JobStatus() != workerpool.Jfinished {
					if !GenVarOpts.finishSignal {
						disposeChan <- j
					}
				} else {
					j.JobDispose()
				}
			}
			close(disposeDone)
		}()
	}

	// wait for all goroutines to finish
	go func() {
		count := 0
		i := 0
		for {
			 verbose("going to wait\n")
			select {
			case <-startedWorkers:
				trace("startedWorkers")
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
				if GenVarOpts.UsePool {
					trace("%v\n", pool.PoolStats())
				}
				if count > 0 {
					info("there are still %d active workers\n", count)
					trace("count = %d and i = %d", count, i)
				} else {
					verbose("count = 0 and i = %d\n", i)
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
						verbose("max depth : %d\n", maxSize)
					}
				} else if strings.HasPrefix(s, "status:") {
					info("%s", s)
				} else {
					GenVarOpts.countGroups++
					out += s
					info("%d\n%s", GenVarOpts.countGroups, s)
					if GenVarOpts.countGroups == GenVarOpts.MaxGroups {
						close(msgs)
						close(collectingDone)
						return
					}
				}
			default:
				time.Sleep(1 * time.Second)
				verbose("%s passed\n", time.Now().Sub(t0))
				if GenVarOpts.UsePool {
					info("%v\n", pool.PoolStats())
				}
				if GenVarOpts.finishSignal {
					info("finishSignal issued, exiting")
					close(msgs)
					return
				}
				debug("current workers %d, max workers %d\n",
					GenVarOpts.currentWorkers, GenVarOpts.maxWorkers)
			}
		}
	}()

	if GenVarOpts.UsePool {
		pool.NewJobQueue(findGroups, findArg{baseGroup, wmap})
	} else {
		go findGroups(findArg{baseGroup, wmap}, nil, nil)
	}

	dur := time.Duration(GenVarOpts.TimeToRun)
	select {
	case <-collectingDone:
		info("required results collected after %s\n", time.Now().Sub(t0))
	case <-time.After(dur * time.Second):
		info("stopped after %s\n", time.Now().Sub(t0))
	}
	GenVarOpts.finishSignal = true

	// pool cleanup
	if GenVarOpts.UsePool {
		ch := make(chan struct{})
		if GenVarOpts.UseJobDispose {
			close(disposeChan)
		}
		pool.StopDispatcher(func(){
			select {
				case <-disposeDone:
			}
			close(ch)
		})
		pool.Dispose()
		<-ch
	}

	info("waiting for waitForWorkers, %d workers\n", GenVarOpts.currentWorkers)
	<-waitForWorkers
	fmt.Printf("exiting... after %s\n", time.Now().Sub(t0))

	fmt.Println(out)
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

func debug(f string, v ...interface{}) { if GenVarOpts.DebugEnabled { fmt.Printf("debug: " + f, v...) } }

func info(f string, v ...interface{}) { if len(GenVarOpts.Verbose) >= 1 && GenVarOpts.Verbose[0] { fmt.Printf("info: "+f, v...) } }

func verbose(f string, v ...interface{}) { if len(GenVarOpts.Verbose) >= 2 && GenVarOpts.Verbose[1] { fmt.Printf("verbose" + f, v...) } }

func trace(f string, v ...interface{}) { if len(GenVarOpts.Verbose) >= 3 && GenVarOpts.Verbose[2] { fmt.Printf("trace" + f, v...) } }
