package cvc

import (
	"fmt"
	"strings"
	"testing"
)

func TestCVC(t *testing.T) {
	print("running TestCVC\n")
}

func TestCVCword(t *testing.T) {
	w := NewCVCWord("X", "E", "Z", 55)

	expected := fmt.Sprintf("c[%s]:v[%s]:c[%s] [%s:%d]",
		w.c1, w.v, w.c2, w.actword, w.freq)
	actual := fmt.Sprintf("%s", w.DumpString())
	if expected != actual {
		t.Errorf("cvcword dump is not in the proper format: expected '%s', actual '%s'",
			expected, actual)
	}
	//t.Fatal("abcd", 1234, "rr55")
	zStr := fmt.Sprint(w)
	if zStr != "XEZ" {
		//out := fmt.Sprintf("%T %v, %T %v", zStr, zStr, w.word, w.word)
		t.Errorf("%T %v, %T %v", zStr, zStr, w.actword, w.actword)
		// t.Error("args", zStr, w.word)
	}
}

func prepareTestData() ([]string, []*CvcWord) {
	words := []string{
		"AAB", // 1
		"CED", // 2
		"FIG", // 3
		"JOK", // 4
		"LUM", // 5
		"NAP", // 6
		"QER", // 7
		"SIT", // 8
		"VOW", // 9
		"XUY", // 10
		"RAZ", // 11
		"BAG", // 12
		"KEB", // 13
	}
	cvcwords := []*CvcWord{
		NewCVCWord(words[0][0:1], words[0][1:2], words[0][2:3], 9),
		NewCVCWord(words[1][0:1], words[1][1:2], words[1][2:3], 19),
		NewCVCWord(words[2][0:1], words[2][1:2], words[2][2:3], 29),
		NewCVCWord(words[3][0:1], words[3][1:2], words[3][2:3], 39),
		NewCVCWord(words[4][0:1], words[4][1:2], words[4][2:3], 49),
		NewCVCWord(words[5][0:1], words[5][1:2], words[5][2:3], 59),
		NewCVCWord(words[6][0:1], words[6][1:2], words[6][2:3], 69),
		NewCVCWord(words[7][0:1], words[7][1:2], words[7][2:3], 79),
		NewCVCWord(words[8][0:1], words[8][1:2], words[8][2:3], 89),
		NewCVCWord(words[9][0:1], words[9][1:2], words[9][2:3], 99),
		NewCVCWord(words[10][0:1], words[10][1:2], words[10][2:3], 109),
		NewCVCWord(words[11][0:1], words[11][1:2], words[11][2:3], 119),
		NewCVCWord(words[12][0:1], words[12][1:2], words[12][2:3], 129),
	}
	return words, cvcwords
}

func TestCVCListAsString(t *testing.T) {
	words, cws := prepareTestData()
	w1 := cws[0]
	w2 := cws[1]
	w3 := cws[2]

	// fmt.Printf("%v: %p\n", w1, w1)

	expected := words[0] + " " + words[1] + " " + words[2]
	wl := &CvcList{}

	if strings.Compare("", fmt.Sprintf("%s", wl)) != 0 {
		t.Errorf("list is not empty")
	}
	if strings.Compare("", fmt.Sprintf("%s", wl.StringWithFreq())) != 0 {
		t.Errorf("list is not empty")
	}
	*wl = append(*wl, w1, w2, w3)
	// fmt.Printf("%v: %p\n", (*wl)[0], *(&(*wl)[0]))
	wl_ptr_w1 := *(&(*wl)[0])
	if wl_ptr_w1 != w1 {
		t.Errorf("test data pointer for word %s (0x%p) is different "+
			"from list word %s pointer (0x%p)", w1, w1, (*wl)[0], wl_ptr_w1)
	}

	// manual chekcing pointers are the same
	// wl2 := &CvcList{}
	// *wl2 = append(*wl2, (*wl)[0])
	// fmt.Printf("%v: %p\n", (*wl2)[0], *(&(*wl2)[0]))

	if wl.asString() != expected {
		msg := fmt.Sprintf("CVCList as string malformed: "+
			"expected [%s], actual [%s]", expected, wl.asString())
		t.Error(msg)
	}

	if len(wl.dump()) != len(strings.SplitN(wl.String(), " ", -1)) {
		t.Error("dump and as string do not correlate", wl.dump(),
			wl.String())
	}
	//fmt.Println(strings.SplitN(wl.String(), ", ", -1))
	//fmt.Println(wl.dump())

	wldump := fmt.Sprintf((*wl)[0].dumpString())
	dumpFormat := fmt.Sprintf("c[%s]:v[%s]:c[%s] [%s:%d]",
		w1.c1, w1.v, w1.c2, w1.actword, w1.freq)
	if wldump != dumpFormat {
		t.Errorf("cvcword dumping '%s' is not in the proper "+
			"format '%s'", wldump, dumpFormat)
	}
}

func TestCVCcontain(t *testing.T) {
	words, cws := prepareTestData()
	w1 := cws[0]
	w2 := cws[1]
	w3 := cws[2]
	w4 := cws[3]

	var wlist CvcList
	wlist = append(wlist, w1, w2, w3)

	if wlist.contain(w4) {
		t.Error(wlist)
		t.Errorf("list %s should not contain %s", wlist, w4)
	}

	if !wlist.contain(w3) {
		t.Error(wlist)
		t.Errorf("list %s should contain %s", wlist, w4)
	}

	if wlist.countCVCWord(w3) > 1 {
		t.Errorf("list %s contain more then one occurence of %s",
			wlist, w3)
	}

	wlist = append(wlist, w3)
	if wlist.countCVCWord(w3) != 2 {
		t.Errorf("list %s does not contain two occurence of %s",
			wlist, w3)
	}

	wlist2 := wlist.CopyList()
	expected := words[0] + " " + words[1] + " " + words[2] + " " + words[2]
	if wlist2.asString() != expected {
		msg := fmt.Sprintf("copy of CVCList as string malformed: "+
			"expected [%s], actual [%s]", expected, wlist2.asString())
		t.Error(msg)
	}
}

func TestCvcSetSimple(t *testing.T) {
	_, cws := prepareTestData()

	set := NewSetLimit(1)

	set.AddWord(cws[0])
	if added, _ := set.AddWord(cws[1]); added == true {
		t.Errorf(`cvcword %s, should not be joined to set
			limited to 1 element %s`, cws[1], set)
	}

}

func TestCvcSetLimit(t *testing.T) {
	_, cws := prepareTestData()

	defer func() {
		panic_msg := "Set not support more then 10 elements"
		if r := recover(); r != nil {
			// fmt.Printf("recover : %T\n", r)
			// fmt.Printf("recover other : [%v]\n", r)
			// fmt.Printf("panic message : [%v]\n", panic_msg)
			if r != panic_msg {
				t.Errorf(`limit support for more then 10 elements
					is not implelemnted`)
			}
		}
	}()
	set := NewSetLimit(11)
	set.AddWord(cws[0])
}

func TestCvcSetFreq(t *testing.T) {
	_, cws := prepareTestData()

	set := NewSetLimitFreq(4, 40, 2)
	set.AddWord(cws[0])
	set.AddWord(cws[1])
	set.AddWord(cws[5])
	if added, _ := set.AddWord(cws[3]); added == true {
		t.Errorf(`cvcword %s, freq %d,
			should not be joined to set %s with %d words of freq %d`,
			cws[3], cws[3].freq, set.StringWithFreq(), set.freqabove,
			set.freqcutoff)
	}

	set.AddWord(cws[6])
	testStringWithFreq := fmt.Sprintf("[%s:%d, %s:%d, %s:%d, %s:%d]",
		cws[0], cws[0].freq,
		cws[1], cws[1].freq,
		cws[5], cws[5].freq,
		cws[6], cws[6].freq,
	)
	if set.StringWithFreq() != testStringWithFreq {
		t.Errorf("set %s, is not as test string %s", set.StringWithFreq(),
			testStringWithFreq)
	}

	set2 := NewSetLimitFreq(2, 40, 1)
	set2.AddWord(cws[5])
	if added, _ := set2.AddWord(cws[6]); added == true {
		t.Errorf(`cvcword %s, freq %d,
			should not be joined to set %s with %d words of freq %d`,
			cws[6], cws[6].freq, set.StringWithFreq(), set.freqabove,
			set.freqcutoff)
	}
}

func TestCvcSet(t *testing.T) {
	words, cws := prepareTestData()
	w1 := cws[0]
	w2 := cws[1]
	w3 := cws[2]
	w4 := cws[3]
	w5 := cws[4]
	w6 := cws[5]
	w7 := cws[6]
	w8 := cws[7]
	w9 := cws[8]
	w10 := cws[9]
	wbad1 := cws[10]
	wbad2 := cws[11]
	wbad3 := cws[12]

	set := NewSet()

	set.AddWord(w1)
	if added, _ := set.AddWord(wbad2); added == true {
		t.Errorf("cvcword %s, should not be joined to set with %s",
			wbad1, set)
	}
	if added, _ := set.AddWord(wbad3); added == true {
		t.Errorf("cvcword %s, should not be joined to set with %s",
			wbad2, set)
	}
	set.AddWord(w2)
	set.AddWord(w3)
	set.AddWord(w4)
	set.AddWord(w5)
	set.AddWord(w6)
	if added, _ := set.AddWord(wbad1); added == true {
		t.Errorf("cvcword %s, should no be joined set already have 2 "+
			"vowels %s", wbad1, set)
	}
	set.AddWord(w7)
	set.AddWord(w8)
	set.AddWord(w9)
	set.AddWord(w10)
	if set.count != 10 {
		t.Errorf("set does not contain 10 cvc words %s", set)
	}
	if added, _ := set.AddWord(wbad1); added == true {
		t.Errorf("cvcword %s, should no be joined to set with %s",
			wbad1, set)
	}
	if _, full := set.AddWord(wbad1); full != true {
		t.Errorf("cvcword %s, should no be joined to set with %s",
			wbad1, set)
	}

	set_string := strings.Replace(set.String(), ",", "", -1)
	words_string := fmt.Sprintf("%s", words[0:10])

	if set_string != words_string {
		t.Errorf("set content: %s, words list: %s\n",
			set_string, words_string)
	}

	set2 := set.CopySet()
	set2_string := strings.Replace(set2.String(), ",", "", -1)

	if set2_string != words_string {
		t.Errorf("copy of set content: %s, words list: %s\n",
			set2_string, words_string)
	}
}

func TestCvcGroupSetSimple(t *testing.T) {
	_, cws := prepareTestData()

	group := NewGroupSetLimit(2, 2)

	group.AddWord(cws[0])
	group.AddWord(cws[1])
	group.AddWord(cws[2])
	group.AddWord(cws[3])
	// fmt.Printf("group content: %s", group)
	if added, _ := group.AddWord(cws[4]); added == true {
		t.Errorf("cvcword %s, should not be joined to "+
			"group with %s", cws[4], group)
	}
	// fmt.Printf("group content: %s", group)
}

func TestCvcGroupSet(t *testing.T) {
	_, cws := prepareTestData()

	group := NewGroupSetLimit(2, 2)
	group.AddWord(cws[0])
	group.AddWord(cws[1])
	if added, _ := group.AddWord(cws[0]); added == true {
		t.Errorf("cvcword %s, should not be joined to "+
			"group with %s", cws[0], group)
	}
	group.AddWord(cws[2])
	group.AddWord(cws[3])

	testString := fmt.Sprintf("\n\t[%s, %s]\n\t[%s, %s]\n", cws[0], cws[1], cws[2], cws[3])
	if group.String() != testString {
		t.Errorf("group '%s', is not as test string '%s'", group.String(),
			testString)
	}

	testStringWithFreq := fmt.Sprintf("\n\t1:[%s:%d, %s:%d]\n\t2:[%s:%d, %s:%d]\n",
		cws[0], cws[0].freq,
		cws[1], cws[1].freq,
		cws[2], cws[2].freq,
		cws[3], cws[3].freq,
	)
	if group.StringWithFreq() != testStringWithFreq {
		t.Errorf("group '%s', is not as test string '%s'", group.StringWithFreq(),
			testStringWithFreq)
	}

	group2 := group.CopyCvcGroupSet()
	if group2.StringWithFreq() != testStringWithFreq {
		t.Errorf("copy group '%s', is not as test string '%s'", group2.StringWithFreq(),
			testStringWithFreq)
	}
}

func TestCvcGroupFreq(t *testing.T) {
	_, cws := prepareTestData()

	cws[0].freq = 17
	cws[1].freq = 21
	cws[2].freq = 4
	cws[3].freq = 105
	cws[4].freq = 11
	cws[5].freq = 9
	cws[6].freq = 29
	cws[7].freq = 309
	cws[8].freq = 4444

	freq_cutoff := 20
	freq_above := 1
	group := NewGroupSetLimitFreq(3, 2, freq_cutoff, freq_above)

	group.AddWord(cws[1])
	// test 2nd word lower then cutoff
	if added, _ := group.AddWord(cws[3]); added {
		t.Errorf("frequency mismatch: cutoff: %d, above cutoff %d: word '%s' should be added to group '%s'",
			freq_cutoff, freq_above, cws[3].dumpString(), group.StringWithFreq())
	}
	group.AddWord(cws[0])
	// test existing word in previous set
	if added, _ := group.AddWord(cws[1]); added {
		t.Errorf("frequency mismatch: cutoff: %d, above cutoff %d: word '%s' should be added to group '%s'",
			freq_cutoff, freq_above, cws[1].dumpString(), group.StringWithFreq())
	}
	group.AddWord(cws[2])
	// test 2nd word lower then cutoff
	if added, _ := group.AddWord(cws[4]); added {
		t.Errorf("frequency mismatch: cutoff: %d, above cutoff %d: word '%s' should be added to group '%s'",
			freq_cutoff, freq_above, cws[4].dumpString(), group.StringWithFreq())
	}
	group.AddWord(cws[3])

	group.AddWord(cws[6])
	// test 2nd word higher then cutoff
	if added, _ := group.AddWord(cws[7]); added {
		t.Errorf("frequency mismatch: cutoff: %d, above cutoff %d: word '%s' should be added to group '%s'",
			freq_cutoff, freq_above, cws[7].dumpString(), group.StringWithFreq())
	}
	group.AddWord(cws[5])

	if _, full := group.AddWord(cws[8]); !full {
		t.Errorf("group '%s' is full", group.StringWithFreq())
	}

	if group.CurrentSize() != group.MaxSize() {
		t.Errorf("group count %f is not fit with group max count %f",
			group.CurrentSize(), group.MaxSize())
	}
}

func TestCvcMap(t *testing.T) {
	_, cws := prepareTestData()

	newmap := NewCvcWordMap()

	newmap.AddWord(cws[0])
	if newmap.AddWord(cws[0]) {
		t.Errorf("cvcword %s, should not be in map %s", cws[0], newmap)
	}
	testString := fmt.Sprintf("%s:1", cws[0])
	if testString != newmap.String() {
		t.Errorf("map '%s', is not as test string '%s'", newmap, testString)
	}

	if newmap.Size() != 1 {
		t.Errorf("map size %d is incorrect %d", newmap.Size(), 1)
	}
	newmap.AddWord(cws[1])

	copymap := newmap.CopyCvcWordMap()
	newmap.AddWord(cws[2])

	testString2 := fmt.Sprintf("%s:1, %s:1", cws[0], cws[1])
	if copymap.String() != testString2 {
		t.Errorf("map copy content '%s' not identical to expected '%s'", copymap, testString2)
	}

	tmap := *newmap.GetCm()
	tmapcopy := *copymap.GetCm()

	if len(tmap) == len(tmapcopy) {
		t.Errorf("copy of word map len '%d' and original word map len '%d' "+
			"should not be the same ", len(tmapcopy), len(tmap))
	}

	copymap.DelWord(cws[0])
	if copymap.DelWord(cws[0]) == true {
		t.Errorf("word '%s' should not be in map", cws[0])
	}
	act := fmt.Sprintf("%s", copymap)
	exp := fmt.Sprintf("%s:1", cws[1])
	if act != exp {
		t.Errorf("content of map '%s' is not as expected '%s'", act, exp)
	}

	// fmt.Printf("%s\n", newmap)
}
