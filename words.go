package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilwo/words/cvc"
)

var globalInfo struct {
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
}

var consonents, vowels map[string]int

var done = make(chan bool)
var stop = make(chan bool)
var msgs = make(chan string, 100)
var status = make(chan string, 100)

func findGroups(group *cvc.CvcGroupSet, wordmap *cvc.CvcWordMap) {
	defer func() {
		recover()
		// z := recover()
		// fmt.Println("recovered from %s", z)
		return
	}()
	exit := false
	zmap := *wordmap.GetCm()
	for k := range zmap {
		select {
		case <-done:
			exit = true
		default:
			if added, full := group.AddWord(k); full == true {
				msg := fmt.Sprintf("group completed\n%s", group.StringWithFreq())
				msgs <- msg
			} else if added {
				wordmap.DelWord(k)
				go findGroups(group.CopyCvcGroupSet(), wordmap.CopyCvcWordMap())
			}
		}
		if exit {
			return
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
		"input consonent file \"%s\", \noutput file \"%s\", \nfilter file \"%s\"\n",
		globalInfo.maxGroups,
		globalInfo.maxSets,
		globalInfo.maxWords,
		globalInfo.freqCutoff,
		globalInfo.freqWordsPerLineAboveCutoff,
		globalInfo.inWordsFile,
		globalInfo.inVowelFile,
		globalInfo.inConsonentFile,
		globalInfo.outResultFile,
		globalInfo.filterFiles)

	consonents = getMap(globalInfo.inConsonentFile)
	vowels = getMap(globalInfo.inVowelFile)
	fmt.Printf("consonents: \n%s\n", getOrderedMapString(consonents))
	fmt.Printf("vowels: \n%s\n", getOrderedMapString(vowels))

	wmap := getWordsMap(globalInfo.inWordsFile)
	fmt.Printf("map size: %d\ncontent:\n%s\n", wmap.Size(), wmap)

	baseGroup := cvc.NewGroupSetLimitFreq(
		globalInfo.maxSets,
		globalInfo.maxWords,
		globalInfo.freqCutoff,
		globalInfo.freqWordsPerLineAboveCutoff)

	// status collector
	go func() {
		for {
			var s string
			select {
			case s := <-status:
				fmt.Println(s)
			default:
			}
			if s == "end messages" {
				break
			}
		}
	}()

	// group collector
	go func() {
		count := 1
		for s := range msgs {
			if s == "end messages" {
				close(stop)
				return
			}
			fmt.Printf("%d\n%s", count, s)
			count++
		}
	}()
	go findGroups(baseGroup, wmap)

	time.Sleep(30 * time.Second)
	close(done)
	msgs <- "end messages"
	status <- "end messages"
	<-stop

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
