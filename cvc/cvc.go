package cvc

import (
	"fmt"
	"sort"
	"strings"
)

//type stringable interface {
//	asString()
//	dump()
//}

// ***************************************
//           CvcWord
// ***************************************

// CvcWord - consonent/vowel/consonent actword bundle strucrt
//  contain frequency for this word in the usage of the word
type CvcWord struct {
	c1      string
	v       string
	c2      string
	actword string
	freq    int
}

// NewCVCWord creating new CVC Word from given elements
func NewCVCWord(c1 string, v string, c2 string, freq int) *CvcWord {
	var w CvcWord
	w.c1 = c1
	w.v = v
	w.c2 = c2
	w.freq = freq
	w.actword = c1 + v + c2
	return &w
}

func (w *CvcWord) dumpString() string {
	return fmt.Sprintf("c[%s]:v[%s]:c[%s] [%s:%d]",
		w.c1, w.v, w.c2, w.actword, w.freq)
}

func (w *CvcWord) DumpString() string {
	return w.dumpString()
}

func (w *CvcWord) String() string {
	return w.actword
}

// ***************************************
//           CvcList
// ***************************************

// CvcList ...
type CvcList []*CvcWord

func (wlist *CvcList) contain(cw *CvcWord) bool {
	for _, e := range *wlist {
		if e.actword == cw.actword {
			return true
		}
	}
	return false
}

func (wlist *CvcList) countCVCWord(cw *CvcWord) int {
	z := 0
	for _, e := range *wlist {
		if e.actword == cw.actword {
			z++
		}
	}
	return z
}

func (wlist *CvcList) dump() []string {
	var out []string
	for _, e := range *wlist {
		out = append(out, e.actword)
	}
	return out
}

func (wlist *CvcList) asString() string {
	var out string
	for _, e := range *wlist {
		out += e.actword + " "
	}
	return out[0 : len(out)-1]
}

func (wlist *CvcList) asStringWithFreq() string {
	var out string
	for _, e := range *wlist {
		out += fmt.Sprintf("%s:%d ", e, e.freq)
	}
	return out[0 : len(out)-1]
}

// String: for formatting purposes of CvcList
func (wlist *CvcList) String() string {
	if len(*wlist) == 0 {
		return ""
	}
	return "[" + strings.Replace(wlist.asString(), " ", ", ", -1) + "]"
}

func (wlist *CvcList) StringWithFreq() string {
	if len(*wlist) == 0 {
		return ""
	}
	return "[" + strings.Replace(wlist.asStringWithFreq(), " ", ", ", -1) + "]"
}

func (wlist *CvcList) CopyList() *CvcList {
	newlist := &CvcList{}
	for _, e := range *wlist {
		*newlist = append(*newlist, e)
	}
	return newlist
}

// func (wlist *CvcList) GetConsFromList() map[string]int {
// 	cmap := make
//
// }

// ***************************************
//           CvcList
// ***************************************

// CvcSet
type CvcSet struct {
	list       CvcList
	cMap       map[string]int
	vMap       map[string]int
	count      int
	setlimit   int
	freqcutoff int
	freqabove  int
}
type CvcSetList []*CvcSet

// NewSet : return new set with limit of 10 elements cvcwords
func NewSet() *CvcSet {
	var newset *CvcSet = &CvcSet{
		list:       CvcList{},
		cMap:       make(map[string]int, 20), // consonent map
		vMap:       make(map[string]int, 5),  // vowel map
		count:      0,
		setlimit:   10,
		freqcutoff: 0,
		freqabove:  0,
	}
	return newset
}

func NewSetLimit(setlimit int) *CvcSet {
	newset := NewSet()
	if setlimit > newset.setlimit {
		panic("Set not support more then 10 elements")
	}
	newset.setlimit = setlimit
	return newset
}

func NewSetLimitFreq(setlimit, fcutoff, fabove int) *CvcSet {
	newset := NewSet()
	newset.setlimit = setlimit
	newset.freqcutoff = fcutoff
	newset.freqabove = fabove
	return newset
}

func (wset *CvcSet) String() string {
	return wset.list.String()
}

func (wset *CvcSet) StringWithFreq() string {
	return wset.list.StringWithFreq()
}

func (wset *CvcSet) freqCheckOk(w *CvcWord) bool {
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

func (wset *CvcSet) AddWord(w *CvcWord) (added bool, full bool) {
	if wset.count == wset.setlimit {
		return false, true
	}
	if _, c1_already := wset.cMap[w.c1]; c1_already {
		// print("%s already in set", w.c1)
		return false, false
	}

	if _, c2_already := wset.cMap[w.c2]; c2_already {
		// print("%s already in set", w.c2)
		return false, false
	}

	if v_count, v_exist := wset.vMap[w.v]; v_exist {
		if v_count > 1 {
			return false, false
		}
	} else {
		wset.vMap[w.v] = 0
	}

	if !wset.freqCheckOk(w) {
		return false, false
	}

	wset.cMap[w.c1] = 1
	wset.cMap[w.c2] = 1
	wset.vMap[w.v] += 1
	wset.count += 1
	wset.list = append(wset.list, w)

	if wset.count == wset.setlimit {
		return true, true
	}
	return true, false
}

func (wset *CvcSet) CopySet() *CvcSet {
	newset := NewSetLimitFreq(
		wset.setlimit, wset.freqcutoff, wset.freqabove)
	for k, v := range wset.cMap {
		newset.cMap[k] = v
	}
	for k, v := range wset.vMap {
		newset.vMap[k] = v
	}
	for _, e := range wset.list {
		newset.list = append(newset.list, e)
	}
	newset.count = wset.count
	return newset
}

type CvcGroupSet struct {
	list        CvcSetList
	count       int
	current     int
	grouplimit  int
	persetlimit int
	freqcutoff  int
	freqabove   int
}

func NewGroupSetLimit(grouplimit, setlimit int) *CvcGroupSet {
	var newgroup *CvcGroupSet = &CvcGroupSet{
		list:        CvcSetList{},
		count:       0,
		current:     0,
		grouplimit:  grouplimit,
		persetlimit: setlimit,
		freqcutoff:  0,
		freqabove:   0,
	}
	return newgroup
}

func NewGroupSetLimitFreq(grouplimit, setlimit, fcutoff, fabove int) *CvcGroupSet {
	newgroup := NewGroupSetLimit(grouplimit, setlimit)
	newgroup.freqcutoff = fcutoff
	newgroup.freqabove = fabove
	return newgroup
}

func (wg *CvcGroupSet) String() string {
	var out string = string("\n")
	for _, set := range wg.list {
		// fmt.Printf("testing %d\n", i)
		out += fmt.Sprintf("\t%s\n", set.String())
	}
	return out
}

func (wg *CvcGroupSet) StringWithFreq() string {
	var out string = string("\n")
	for _, set := range wg.list {
		// fmt.Printf("testing %d\n", i)
		out += fmt.Sprintf("\t%s\n", set.StringWithFreq())
	}
	return out
}

func (wg *CvcGroupSet) Count() int {
	count := wg.count
	if wg.count == wg.grouplimit {
		count--
	}
	return count*wg.persetlimit + wg.list[wg.current].count
}
func (wg *CvcGroupSet) MaxCount() int {
	return wg.grouplimit * wg.persetlimit
}

func (wg *CvcGroupSet) AddWord(w *CvcWord) (added bool, full bool) {
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
		wg.count += 1
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

func (wg *CvcGroupSet) CopyCvcGroupSet() *CvcGroupSet {
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

type CvcWordMap struct {
	cm   map[*CvcWord]int
	keys CvcList
}

func (wmap *CvcWordMap) GetCm() *map[*CvcWord]int {
	return &wmap.cm
}

func NewCvcWordMap() *CvcWordMap {
	var newmap *CvcWordMap = &CvcWordMap{
		cm: make(map[*CvcWord]int),
		// TODO: check if we need to initialize the key var ?
		// keys: make([]*CvcWord, 1),
	}
	return newmap
}

func (wmap *CvcWordMap) CopyCvcWordMap() *CvcWordMap {
	newmap := NewCvcWordMap()
	for k, v := range wmap.cm {
		newmap.cm[k] = v
	}
	newmap.keys = make(CvcList, len(wmap.keys))
	copy(newmap.keys, wmap.keys)

	// the following line create error in the TestCvcMap
	// newmap.keys = append(CvcList{}, wmap.keys[0])
	return newmap
}

func (wmap *CvcWordMap) AddWord(w *CvcWord) bool {
	if _, w_already := wmap.cm[w]; w_already {
		// print(w, " already in pool")
		return false
	}
	wmap.keys = append(wmap.keys, w)
	wmap.cm[w] = 1
	return true
}

func (wmap *CvcWordMap) DelWord(w *CvcWord) bool {
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
	return true
}

func (wmap *CvcWordMap) String() string {
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

func (wmap *CvcWordMap) Size() int {
	return len(wmap.cm)
}

// // CvcPool ...
// type CvcPool map[string]map[string]CvcList
