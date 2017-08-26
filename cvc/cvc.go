package cvc

import (
	"fmt"
	"sort"
	"strings"
)

// ***************************************
//           Word
// ***************************************

// Word - consonant/vowel/consonant actword bundle strucrt
//  contain frequency for this word in the usage of the word
type Word struct {
	c1      string
	v       string
	c2      string
	actword string
	freq    int
}

// NewWord creating new CVC Word from given elements
func NewWord(c1 string, v string, c2 string, freq int) *Word {
	w := new(Word)
	w.c1 = c1
	w.v = v
	w.c2 = c2
	w.freq = freq
	w.actword = c1 + v + c2
	return w
}

func (w *Word) dumpString() string {
	return fmt.Sprintf("c[%s]:v[%s]:c[%s] [%s:%d]",
		w.c1, w.v, w.c2, w.actword, w.freq)
}

// DumpString : TODO: fill me
func (w *Word) DumpString() string {
	return w.dumpString()
}

func (w *Word) String() string {
	return w.actword
}

// ***************************************
//           WordList
// ***************************************

// WordList : TODO: fill me
type WordList []*Word

func (wlist *WordList) contain(cw *Word) bool {
	for _, e := range *wlist {
		if e.actword == cw.actword {
			return true
		}
	}
	return false
}

func (wlist *WordList) countCVCWord(cw *Word) int {
	z := 0
	for _, e := range *wlist {
		if e.actword == cw.actword {
			z++
		}
	}
	return z
}

func (wlist *WordList) dump() []string {
	var out []string
	for _, e := range *wlist {
		out = append(out, e.actword)
	}
	return out
}

func (wlist *WordList) asString() string {
	var out string
	for _, e := range *wlist {
		out += e.actword + " "
	}
	return out[0 : len(out)-1]
}

func (wlist *WordList) asStringWithFreq() string {
	var out string
	for _, e := range *wlist {
		out += fmt.Sprintf("%s:%d ", e, e.freq)
	}
	return out[0 : len(out)-1]
}

func (wlist *WordList) String() string {
	if len(*wlist) == 0 {
		return ""
	}
	return "[" + strings.Replace(wlist.asString(), " ", ", ", -1) + "]"
}

// StringWithFreq : TODO: fill me
func (wlist *WordList) StringWithFreq() string {
	if len(*wlist) == 0 {
		return ""
	}
	return "[" + strings.Replace(wlist.asStringWithFreq(), " ", ", ", -1) + "]"
}

// CopyList : TODO: fill me
func (wlist *WordList) CopyList() *WordList {
	newlist := &WordList{}
	for _, e := range *wlist {
		*newlist = append(*newlist, e)
	}
	return newlist
}

// ***************************************
//           WordSet
// ***************************************

type cbundle struct {
	consonant string
	exist     bool
}
type vbundle struct {
	vowel string
	count int
}

// WordSet : TODO: fill me
type WordSet struct {
	list       WordList
	cMap       []cbundle
	vMap       []vbundle
	count      int
	setlimit   int
	freqcutoff int
	freqabove  int
}

// WordSetList : TODO: fill me
type WordSetList []*WordSet

// DumpSet : TODO: fill me
func (wset *WordSet) DumpSet() string {
	return fmt.Sprintf("list:\n%s\n"+
		"consonants:\n%v\n"+
		"vowels:\n%v\n"+
		"count:%d\n"+
		"setlimit:%d\n"+
		"freqcutoff:%d\n"+
		"freqabove:%d\n",
		wset.list.asStringWithFreq(),
		wset.cMap,
		wset.vMap,
		wset.count,
		wset.setlimit,
		wset.freqcutoff,
		wset.freqabove)
}

// NewSet : return new set with limit of 10 elements cvcwords
func NewSet() *WordSet {
	return newSetConfigurable(20, 5, 10)
}

// newSetConfigurable : return new set with configurable limits
func newSetConfigurable(consonantCount, vowelCount, setLimit int) *WordSet {
	newset := new(WordSet)
	newset.cMap = make([]cbundle, consonantCount)
	newset.vMap = make([]vbundle, vowelCount)
	newset.setlimit = setLimit
	return newset
}

// NewSetLimit : TODO: fill me
func NewSetLimit(setlimit int) *WordSet {
	newset := NewSet()
	if setlimit > newset.setlimit {
		panic("Set not support more then 10 elements")
	}
	newset.setlimit = setlimit
	return newset
}

// NewSetLimitFreq : TODO: fill me
func NewSetLimitFreq(setlimit, fcutoff, fabove int) *WordSet {
	newset := NewSet()
	newset.setlimit = setlimit
	newset.freqcutoff = fcutoff
	newset.freqabove = fabove
	return newset
}

func (wset *WordSet) String() string {
	return wset.list.String()
}

// StringWithFreq : TODO: fill me
func (wset *WordSet) StringWithFreq() string {
	return wset.list.StringWithFreq()
}

func (wset *WordSet) freqCheckOk(w *Word) bool {
	if wset.freqcutoff == 0 {
		return true
	}

	var acount, bcount int = 0, 0
	if w.freq > wset.freqcutoff {
		acount++
	} else {
		bcount++
	}

	for _, e := range wset.list {
		if e.freq >= wset.freqcutoff {
			acount++
		} else {
			bcount++
		}
	}
	if acount > wset.freqabove {
		return false
	} else if acount+bcount == wset.setlimit && acount < wset.freqabove {
		return false
	}

	return true
}

// AddWord : TODO: fill me
func (wset *WordSet) AddWord(w *Word) (added bool, full bool) {
	if wset.count == wset.setlimit {
		return false, true
	}
	// check consonant validity : do not appear already in the list of cvc words
	var fc int
	for i, e := range wset.cMap {
		fc = i
		if e.consonant == "" {
			break
		}
		if w.c1 == e.consonant {
			if e.exist {
				return false, false
			}
		}
		if w.c2 == e.consonant {
			if e.exist {
				return false, false
			}
		}
	}

	// check vowel validity : do not appear more then twice
	var fv int
	for i, e := range wset.vMap {
		fv = i
		if e.vowel == "" {
			break
		}
		if w.v == e.vowel {
			if e.count > 1 { // if its already 2 we dont want to add another one
				return false, false
			}
			fv = i
			break
		}
	}

	if !wset.freqCheckOk(w) {
		return false, false
	}

	// update the counters for the consonants
	wset.cMap[fc] = cbundle{w.c1, true}
	wset.cMap[fc+1] = cbundle{w.c2, true}

	// update the counter for the vowel
	if wset.vMap[fv].vowel == "" {
		wset.vMap[fv] = vbundle{w.v, 1}
	} else {
		wset.vMap[fv].count++
	}

	// update the cvc list counter
	wset.count++
	// and add to the list
	wset.list = append(wset.list, w)

	if wset.count == wset.setlimit {
		return true, true
	}
	return true, false
}

// CopySet : TODO: fill me
func (wset *WordSet) CopySet() *WordSet {
	newset := NewSetLimitFreq(
		wset.setlimit, wset.freqcutoff, wset.freqabove)
	copy(newset.cMap, wset.cMap)
	copy(newset.vMap, wset.vMap)
	newset.list = append(WordList{}, wset.list...)
	newset.count = wset.count
	return newset
}

// ***************************************
//           GroupSet
// ***************************************

// GroupSet : TODO: fill me
type GroupSet struct {
	list        WordSetList
	count       int // current count of WordSet in group
	current     int // current (not filled) WordSet in group
	grouplimit  int // max amount of WordSet in group
	persetlimit int // max amount of Words in each WordSet in the group
	freqcutoff  int // frequency threshold
	freqabove   int // number of elements required to be above threshold
}

// DumpGroup : TODO: fill me
func (wg *GroupSet) DumpGroup() string {
	var out = string("\n")
	for i, set := range wg.list {
		// fmt.Printf("testing %d\n", i)
		out += fmt.Sprintf("\t%d:%s\n", i+1, set.DumpSet())
	}
	return fmt.Sprintf("list:\n%s\n"+
		"count:%d\n"+
		"current:%d\n"+
		"grouplimit:%d\n"+
		"persetlimit:%d\n"+
		"freqcutoff:%d\n"+
		"freqabove:%d\n",
		out,
		wg.count,
		wg.current,
		wg.grouplimit,
		wg.persetlimit,
		wg.freqcutoff,
		wg.freqabove)
}

// NewGroupSetLimit : TODO: fill me
func NewGroupSetLimit(grouplimit, setlimit int) *GroupSet {
	newgroup := new(GroupSet)
	newgroup.list = WordSetList{}
	newgroup.grouplimit = grouplimit
	newgroup.persetlimit = setlimit
	return newgroup
}

// NewGroupSetLimitFreq : TODO: fill me
func NewGroupSetLimitFreq(grouplimit, setlimit, fcutoff, fabove int) *GroupSet {
	newgroup := NewGroupSetLimit(grouplimit, setlimit)
	newgroup.freqcutoff = fcutoff
	newgroup.freqabove = fabove
	return newgroup
}

func (wg *GroupSet) String() string {
	var out = string("\n")
	for _, set := range wg.list {
		// fmt.Printf("testing %d\n", i)
		out += fmt.Sprintf("\t%s\n", set.String())
	}
	return out
}

// StringWithFreq : TODO: fill me
func (wg *GroupSet) StringWithFreq() string {
	var out = string("\n")
	for i, set := range wg.list {
		// fmt.Printf("testing %d\n", i)
		out += fmt.Sprintf("\t%d:%s\n", i+1, set.StringWithFreq())
	}
	return out
}

// CurrentSize : TODO: fill me
func (wg *GroupSet) CurrentSize() int {
	count := wg.count
	current_list_count := 0
	if wg.count == wg.grouplimit {
		count--
	}
	if len(wg.list) > 0 {
		current_list_count = wg.list[wg.current].count
	}
	return count*wg.persetlimit + current_list_count
}

// MaxSize : TODO: fill me
func (wg *GroupSet) MaxSize() int {
	return wg.grouplimit * wg.persetlimit
}

// AddWord : TODO: fill me
func (wg *GroupSet) AddWord(w *Word) (added bool, full bool) {
	// fmt.Printf("count: %d\n", wg.count)
	switch {
	case wg.count == wg.grouplimit && wg.list[wg.current].count == wg.persetlimit:
		return false, true
	case wg.count == 0:
		fallthrough
	case wg.list[wg.current].count == wg.persetlimit:
		// fmt.Printf("adding new set\n")
		// wg.list = append(wg.list, NewSetLimit(wg.persetlimit))
		wg.list = append(wg.list,
			NewSetLimitFreq(wg.persetlimit, wg.freqcutoff, wg.freqabove))
		wg.count++
	}
	// fmt.Printf("count: %d\n", wg.count)
	wg.current = wg.count - 1 // count is one bases, current is zero based
	for _, set := range wg.list {
		if set.list.contain(w) {
			return false, false
		}
	}
	added, _ = wg.list[wg.current].AddWord(w)
	return added, false
}

// CopyGroupSet : TODO: fill me
func (wg *GroupSet) CopyGroupSet() *GroupSet {
	newgroup := NewGroupSetLimitFreq(
		wg.grouplimit, wg.persetlimit, wg.freqcutoff, wg.freqabove)

	newgroup.count = wg.count
	newgroup.current = wg.current

	for _, e := range wg.list {
		newset := e.CopySet()
		newgroup.list = append(newgroup.list, newset)
	}

	return newgroup
}

// Checkifavailable : TODO: fill me
func (wg *GroupSet) Checkifavailable(wmap *WordMap) bool {
	if wg.MaxSize()-wg.CurrentSize() > wmap.count {
		fmt.Printf("missing : %d, available %d\n", int(wg.MaxSize())-int(wg.CurrentSize()),
			wmap.count)
		return false
	}

	aboveFreqRequiredCount := wg.freqabove * wg.grouplimit
	belowFreqRequiredCount := wg.grouplimit * wg.persetlimit - aboveFreqRequiredCount

	aboveFreqCurrentCount := wg.count * wg.freqabove
	belowFreqCurrentCount := wg.count * (wg.persetlimit - wg.freqabove)

	aboveFreqMissing := aboveFreqRequiredCount - aboveFreqCurrentCount
	belowFreqMissing := belowFreqRequiredCount - belowFreqCurrentCount

	var aboveAvailableCount, belowAvailableCount int
	for _, v := range *wmap.GetCm() {
		if v > wg.freqcutoff {
			aboveAvailableCount++
		} else {
			belowAvailableCount++
		}
	}
	if belowAvailableCount < belowFreqMissing || aboveAvailableCount < aboveFreqMissing {
		return false
	}

	return true
}

// ***************************************
//           WordMap
// ***************************************

// WordMap : TODO: fill me
type WordMap struct {
	cm    map[*Word]int
	keys  WordList
	count int
}

// GetCm : TODO: fill me
func (wmap *WordMap) GetCm() *map[*Word]int {
	return &wmap.cm
}

// NewWordMap : TODO: fill me
func NewWordMap() *WordMap {
	var newmap = &WordMap{
		cm: make(map[*Word]int),
		// TODO: check if we need to initialize the key var ?
		// keys: make([]*Word, 1),
	}
	return newmap
}

// CopyWordMap TODO: fill me
func (wmap *WordMap) CopyWordMap() *WordMap {
	newmap := NewWordMap()
	for k, v := range wmap.cm {
		newmap.cm[k] = v
	}
	newmap.keys = make(WordList, len(wmap.keys))
	copy(newmap.keys, wmap.keys)
	newmap.count = wmap.count

	// the following line create error in the TestCvcMap
	// newmap.keys = append(WordList{}, wmap.keys[0])
	return newmap
}

// AddWord TODO: fill me
func (wmap *WordMap) AddWord(w *Word) bool {
	if _, w_already := wmap.cm[w]; w_already {
		// print(w, " already in pool")
		return false
	}
	wmap.keys = append(wmap.keys, w)
	wmap.cm[w] = w.freq
	wmap.count++
	return true
}

// DelWord TODO: fill me
func (wmap *WordMap) DelWord(w *Word) bool {
	if _, w_exist := wmap.cm[w]; !w_exist {
		return false
	}
	for i, k := range wmap.keys {
		if k == w {
			wmap.keys = append(wmap.keys[:i], wmap.keys[i+1:]...)
			break
		}
	}
	delete(wmap.cm, w)
	wmap.count--
	return true
}

func (wmap *WordMap) String() string {
	out := ""
	sortedkeys := wmap.keys

	sort.Slice(sortedkeys,
		func(i, j int) bool {
			// the order is A to Z, comparator left < right which result is -1
			return strings.Compare(sortedkeys[i].actword,
				sortedkeys[j].actword) == -1
		})

	// for k, v := range wmap.cm {
	// 	out += fmt.Sprintf("%s:%d, ", k.String(), v)
	// }
	for _, k := range sortedkeys {
		out += fmt.Sprintf("%s:%d, ", k.String(), wmap.cm[k])
	}
	return out[:len(out)-2]
}

// Size TODO: fill me
func (wmap *WordMap) Size() int {
	return len(wmap.cm)
}

