package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gilwo/words/cvc"
)

var consonents, vowels map[string]int

type workert chan struct{}
type msgt chan string

func findGroups(doneCollector chan workert, report msgt, group *cvc.CvcGroupSet, wordmap *cvc.CvcWordMap) {
	done := make(workert)
	defer func() {
		recover()
		// z := recover()
		// fmt.Println("recovered from %s", z)
		return
	}()
	doneCollector <- done
	zmap := *wordmap.GetCm()
	for k, _ := range zmap {
		if added, full := group.AddWord(k); full == true {
			msg := fmt.Sprintf("group completed\n%s", group.StringWithFreq())
			fmt.Println(msg)
			select {
			case <-done:
				// fmt.Println("done issued")
				return
			default:
				report <- msg
			}
		} else if added == false {
			continue
		}
		wordmap.DelWord(k)
		go findGroups(doneCollector, report, group.CopyCvcGroupSet(), wordmap.CopyCvcWordMap())
	}
}

func collectReports2(done, collectingDone workert, report msgt) {
	const reportingDone string = "reporting done"
	for {
		select {
		case msg := <-report:
			if msg == reportingDone {
				fmt.Println("reached end of reports: closing done")
				close(done)
				return
			}
			fmt.Println(msg)
		case <-collectingDone:
			go func() {
				fmt.Println("reached end of reports")
				report <- "reporting done"
			}()
			// default:
		}
	}
}

func main() {

	wmap := getWordsMap()
	fmt.Printf("map size: %d\ncontent:\n%s\n", wmap.Size(), wmap)

	consonents = getMap("consonents.txt")
	vowels = getMap("vowels.txt")

	fmt.Printf("consonents: \n%s\n", getOrderedMapString(consonents))
	fmt.Printf("vowels: \n%s\n", getOrderedMapString(vowels))

	doneGroupCollector := make(chan workert)
	dummyDone := make(workert)
	doneReporter := make(workert)
	doneCollecting := make(workert)
	report := make(msgt)
	g1 := cvc.NewGroupSetLimitFreq(20, 6, 0, 0)

	go collectReports2(doneReporter, doneCollecting, report)
	go findGroups(doneGroupCollector, report, g1, wmap)

	time.Sleep(10 * time.Second)

	go func() { time.Sleep(1 * time.Second); doneGroupCollector <- dummyDone }()
	for d := range doneGroupCollector {
		if d == dummyDone {
			fmt.Println("dummy reached")
			break
		}
		close(d)

	}
	doneCollecting <- struct{}{}
	<-doneReporter
	close(doneGroupCollector)
	close(report)
}

func getOrderedMapString(m map[string]int) string {
	out := ""
	var sortedkeys []string
	for k, _ := range m {
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
	var ret map[string]int = make(map[string]int)
	for _, wf := range getWordsFromFile(mapfile) {
		ret[wf.word] = wf.number
	}
	return ret
}

func getWordsMap() *cvc.CvcWordMap {
	wmap := cvc.NewCvcWordMap()

	for _, wf := range getWordsFromFile("words_list.txt") {
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
