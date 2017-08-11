package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilwo/words/cvc"
)

var globalInfo struct {
	// flagable vars
	maxGroups                   int
	freqCutoff                  int
	freqWordsPerLineAboveCutoff int
	vowelLimit                  int // 2
	maxSets                     int
	maxWords                    int
	outResultFile               string
	inWordsFile                 string
	inConsonentFile             string
	inVowelFile                 string
	filterFiles                 string
	timeToRun                   int
	cpuprofile                  string
	memprofile                  string
	debugEnabled   bool

	// internal vars
	countGroups    int
	finishSignal   bool
	maxWorkers     int
	currentWorkers int
}

func init() {
	const (
		defaultVowelsFile     = "vowels.txt"
		defaultConsonentsFile = "consonents.txt"
		defaultWordsFile      = "words_list.txt"
	)
	flag.IntVar(&globalInfo.maxGroups, "maxres", 20,
		"maximum number of result groups to generate")
	flag.IntVar(&globalInfo.maxSets, "maxset", 15,
		"numbner of sets per group")
	flag.IntVar(&globalInfo.maxWords, "maxwords", 10,
		"numbner of words per set")
	flag.IntVar(&globalInfo.freqCutoff, "freq", 25,
		"frequency cutoff threshold for words, lower is more common")
	flag.IntVar(&globalInfo.freqWordsPerLineAboveCutoff, "freqcutoff", 3,
		"how many words to be above cutoff threshold per line")

	flag.StringVar(&globalInfo.inConsonentFile, "consonents", defaultConsonentsFile,
		"input file name for words list to use for creating the lines groups results")
	flag.StringVar(&globalInfo.inVowelFile, "vowels", defaultVowelsFile,
		"input file name for words list to use for creating the lines groups results")
	flag.StringVar(&globalInfo.inWordsFile, "words", defaultWordsFile,
		"input file name for words list to use for creating the lines groups results")

	flag.StringVar(&globalInfo.filterFiles, "filter", "",
		"input file name for filtered words")
	flag.StringVar(&globalInfo.outResultFile, "output", "",
		"output file for generated results (default when not set words_result.txt)")

	flag.IntVar(&globalInfo.timeToRun, "timeToRun", 30,
		"how much time to run (in seconds), default 30 sec")

	flag.StringVar(&globalInfo.cpuprofile, "cpuprofile", "",
		"enable cpuprofling and save to file")

	flag.StringVar(&globalInfo.memprofile, "memprofile", "",
		"enable memprofling and save to file")

	flag.BoolVar(&globalInfo.debugEnabled, "debug", false,
	"enable debugging information")
}

var consonents, vowels map[string]int

var waitForWorkers = make(chan bool)
var collectingDone = make(chan struct{})
var msgs = make(chan string, 100)
var startedWorkers = make(chan struct{}, 100)
var stoppedWorkers = make(chan struct{}, 100)
var maxSize int = 0

func findGroups(group *cvc.CvcGroupSet, wordmap *cvc.CvcWordMap) {
	defer func() {
		if fail := recover(); fail != nil {
			// fmt.Printf("recovered from %s\n", fail)
		}
		stoppedWorkers <- struct{}{}
		group = nil
		wordmap = nil
		// runtime.GC()
		return
	}()

	startedWorkers <- struct{}{}
	zmap := *wordmap.GetCm()
	if float64(group.CurrentSize())/float64(group.MaxSize()) > float64(0.9) {
		s := fmt.Sprintf("status: reached depth %d of %d\n",
			int(group.CurrentSize()), int(group.MaxSize()))
		msgs <- s
	}
	if group.CurrentSize() > maxSize {
		msgs <- "depth: " + strconv.Itoa(group.CurrentSize())
	}

	for k := range zmap {

		if globalInfo.finishSignal {
			// fmt.Println("finishSignal issued, exiting")
			break
		}
		if globalInfo.countGroups >= globalInfo.maxGroups {
			// fmt.Printf("groups count %d reached max groups %d", globalInfo.countGroups, globalInfo.maxGroups)
			break
		}
		if added, full := group.AddWord(k); full == true {
			msg := fmt.Sprintf("group completed\n%s\n", group.StringWithFreq())
			if globalInfo.debugEnabled {
				msg += group.DumpGroup() + "\n"
			}
			// fmt.Println(msg)
			msgs <- msg
			break
		} else if added {
			wordmap.DelWord(k)
			go findGroups(group.CopyCvcGroupSet(), wordmap.CopyCvcWordMap())
		}
	}
}

func main() {

	flag.Parse()
	if globalInfo.outResultFile == "" {
		globalInfo.outResultFile = "words_result.txt"
	}
	fmt.Printf("\nlooking for max %d groups of %d sets (%d per set), "+
		"with frequency cutoff of %d, %d words above cutoff threshold for each set\n"+
		"using input word file \"%s\", \ninput vowel file \"%s\", \n"+
		"input consonent file \"%s\", \noutput file \"%s\", \nfilter file \"%s\"\n"+
		"running for %d seconds\n",
		globalInfo.maxGroups,
		globalInfo.maxSets,
		globalInfo.maxWords,
		globalInfo.freqCutoff,
		globalInfo.freqWordsPerLineAboveCutoff,
		globalInfo.inWordsFile,
		globalInfo.inVowelFile,
		globalInfo.inConsonentFile,
		globalInfo.outResultFile,
		globalInfo.filterFiles,
		globalInfo.timeToRun)

	if globalInfo.memprofile != "" {
		f, err := os.Create(globalInfo.memprofile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("enable memprofiling, write to '%v'\n", globalInfo.memprofile)
		defer func() {
			pprof.WriteHeapProfile(f)
			f.Close()
			return
		}()
	}

	if globalInfo.cpuprofile != "" {
		f, err := os.Create(globalInfo.cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("enable cpuprofiling, write to '%v'\n", globalInfo.cpuprofile)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	consonents = getMap(globalInfo.inConsonentFile)
	vowels = getMap(globalInfo.inVowelFile)
	fmt.Printf("consonents: %d\n%s\n", len(consonents), getOrderedMapString(consonents))
	fmt.Printf("vowels: %d\n%s\n", len(vowels), getOrderedMapString(vowels))

	wmap := getWordsMap(globalInfo.inWordsFile)
	fmt.Printf("map size: %d\ncontent:\n%s\n", wmap.Size(), wmap)

	// set the base group accoring to the required settings
	baseGroup := cvc.NewGroupSetLimitFreq(
		globalInfo.maxSets,
		globalInfo.maxWords,
		globalInfo.freqCutoff,
		globalInfo.freqWordsPerLineAboveCutoff)

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
				if count > globalInfo.maxWorkers {
					globalInfo.maxWorkers = count
				}
				globalInfo.currentWorkers = count
				i = 0
			case <-stoppedWorkers:
				// println("stoppedWorkers")
				count--
				globalInfo.currentWorkers = count
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
					globalInfo.countGroups++
					fmt.Printf("%d\n%s", globalInfo.countGroups, s)
					if globalInfo.countGroups == globalInfo.maxGroups {
						close(msgs)
						close(collectingDone)
						return
					}
				}
			default:
				time.Sleep(1 * time.Second)
				fmt.Printf("%s passed\n", time.Now().Sub(t0))
				if globalInfo.finishSignal {
					// fmt.Println("finishSignal issued, exiting")
					close(msgs)
					return
				}
				if globalInfo.debugEnabled {
					fmt.Printf("current workers %d, max workers %d\n",
						globalInfo.currentWorkers, globalInfo.maxWorkers)
				}
			}
		}
	}()

	go findGroups(baseGroup, wmap)

	dur := time.Duration(globalInfo.timeToRun)
	select {
	case <-collectingDone:
		fmt.Printf("required results collected after %s\n", time.Now().Sub(t0))
	case <-time.After(dur * time.Second):
		fmt.Printf("stopped after %s\n", time.Now().Sub(t0))
	}
	globalInfo.finishSignal = true

	fmt.Println("waiting for waitForWorkers")
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

func getWordsMap(fname string) *cvc.CvcWordMap {
	wmap := cvc.NewCvcWordMap()

	for _, wf := range getWordsFromFile(fname) {
		var cvcw *cvc.CvcWord
		wfV := string(wf.word[1])
		if _, ok := vowels[wfV]; ok {
			cvcw = cvc.NewCVCWord(
				string(wf.word[0]),
				string(wf.word[1]),
				string(wf.word[2:]),
				wf.number)
		} else {
			cvcw = cvc.NewCVCWord(
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

// var consonents = map[string]int{
// 	"B": 0, "G": 0, "D": 0, "V": 0, "Z": 0, "X": 0, "T": 0,
// 	"J": 0, "K": 0, "L": 0, "M": 0, "N": 0, "S": 0, "P": 0, "F": 0, "TZ": 0,
// 	"R": 0, "SH": 0, "Q": 0, "W": 0, "H": 0}
// var vowels = map[string]int{
// 	"A": 0, "E": 0, "I": 0, "O": 0, "U": 0}
