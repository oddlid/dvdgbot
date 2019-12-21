package leet

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	boolVar bool
	intVar  int
	strVar  string
	tcVar   TimeCode
)

func getData() *ScoreData {
	const c string = "#dvdg"
	sd := newScoreData()
	//sd.get(c)

	fmt.Println("Creating Oddlid")
	o := sd.get(c).get("Oddlid")
	fmt.Println("Creating Tord")
	t := sd.get(c).get("Tord")
	fmt.Println("Creating Snelhest")
	s := sd.get(c).get("Snelhest")
	fmt.Println("Creating bAAAArd")
	b := sd.get(c).get("bAAAArd")

	o.score(10)
	t.score(8)
	s.score(6)
	b.score(4)

	return sd
}

func TestSave(t *testing.T) {
	const fname string = "/tmp/leetdata.json"
	file, err := os.Create(fname)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	sd := getData()
	n, err := sd.save(file)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Saved %d bytes to %q", n, fname)
}

func TestBonusConfigCalc(t *testing.T) {
	var bcs BonusConfigs
	stamps := []string{
		"00000000000",
		"13370000000",
		"01337000000",
		"00133700000",
		"00013370000",
		"00001337000",
		"00000133700",
		"00000013370",
		"00000001337",
		"01000001337",
		"00006661337",
		"00006660000",
		"00001337666",
	}

	exp := 0
	got := 0

	bc1 := BonusConfig{
		SubString:    "1337",
		PrefixChar:   '0',
		UseStep:      true,
		StepPoints:   10,
		NoStepPoints: 0,
	}
	bc2 := BonusConfig{
		SubString:    "666",
		PrefixChar:   '0',
		UseStep:      false,
		StepPoints:   0,
		NoStepPoints: 18,
	}

	bcs.add(bc1)

	for i := 0; i <= 8; i++ {
		exp = i * bc1.StepPoints
		got = bcs.calc(stamps[i])
		if got != exp {
			t.Errorf("Expected %d, got %d from substring %s", exp, got, stamps[i])
		}
		t.Logf("%s gives %d points bonus", stamps[i], got)
	}

	bcs.add(bc2)

	exp = 18
	ts := stamps[11]
	got = bcs.calc(ts)
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

	exp = 28
	ts = stamps[10]
	got = bcs.calc(ts)
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

	exp = 68
	ts = stamps[12]
	got = bcs.calc(ts)
	if got != exp {
		t.Errorf("Expected %d, got %d from substring %s", exp, got, ts)
	} else {
		t.Logf("%s gives %d points bonus", ts, got)
	}

}


// This BM shows that almost all execution time in bonus() goes to
// 2 lines of string formatting...
func BenchmarkStrFmt(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	var str1 string
	var str2 string

	for i := 0; i < b.N; i++ {
		str1 = fmt.Sprintf("%02d%09d", ts.Second(), ts.Nanosecond())
		str2 = fmt.Sprintf("%02d%02d", 13, 37)
	}
	strVar = str1
	strVar = str2
}

func BenchmarkStrIndex(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	tstr := fmt.Sprintf("%02d%09d", ts.Second(), ts.Nanosecond())
	sstr := "1337"
	var ires int

	for i := 0; i < b.N; i++ {
		ires = strings.Index(tstr, sstr)
	}
	intVar = ires
}

//func BenchmarkPrefixedBy(b *testing.B) {
//	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
//	tstr := fmt.Sprintf("%02d%09d", ts.Second(), ts.Nanosecond())
//	sstr := "1337"
//	idx := strings.Index(tstr, sstr)
//	var bres bool
//
//	for i := 0; i < b.N; i++ {
//		bres = pb(tstr, '0', idx)
//	}
//	boolVar = bres
//}

//func BenchmarkBonus(b *testing.B) {
//	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
//	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:13.370000000Z")
//	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:01.337000000Z")
//	//bp := 0
//	var ires int
//	for i := 0; i < b.N; i++ {
//		//bp = bonus(ts)
//		ires = bonus(ts)
//	}
//	intVar = ires
//	//b.Logf("bonus: %d", bp)
//}

func BenchmarkBonusConfigCalc(b *testing.B) {
	var bcs BonusConfigs
	stamps := []string{
		"00000000000",
		"13370000000",
		"01337000000",
		"00133700000",
		"00013370000",
		"00001337000",
		"00000133700",
		"00000013370",
		"00000001337",
		"01000001337",
		"00006661337",
		"00006660000",
		"00001337666",
	}

	bc1 := BonusConfig{
		SubString:    "1337",
		PrefixChar:   '0',
		UseStep:      true,
		StepPoints:   10,
		NoStepPoints: 0,
	}
	bc2 := BonusConfig{
		SubString:    "666",
		PrefixChar:   '0',
		UseStep:      false,
		StepPoints:   0,
		NoStepPoints: 18,
	}

	bcs.add(bc1)
	bcs.add(bc2)

	got := 0

	for _, ts := range stamps {
		for i := 0; i < b.N; i++ {
			got = bcs.calc(ts)
		}
		intVar = got
	}
}

func BenchmarkTryScore(b *testing.B) {
	tm := func(tstr string) time.Time {
		t, _ := time.Parse(time.RFC3339Nano, tstr)
		return t
	}
	sd := newScoreData()
	channel := "#channel"
	nicks := []struct {
		nick string
		ts time.Time
	}{
		{"Odd_01", tm("2019-04-07T13:37:00.000001337Z")},
		{"Odd_02", tm("2019-04-07T13:37:00.000013370Z")},
		{"Odd_03", tm("2019-04-07T13:37:00.000133700Z")},
	}
	c := sd.get(channel)
	var bres bool
	var sres string
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, n := range nicks {
			bres, sres = sd.tryScore(channel, n.nick, n.ts)
		}
		//c.MergeScoresForRound(c.GetScoresForRound())
		c.clearNicksForRound()
	}
	boolVar = bres
	strVar = sres
}

func BenchmarkWithinTimeFrame(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	var bres bool
	var tcres TimeCode
	for i := 0; i < b.N; i++ {
		bres, tcres = withinTimeFrame(ts)
	}
	boolVar = bres
	tcVar = tcres
}

func BenchmarkTimeFrame(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	var result TimeCode
	for i := 0; i < b.N; i++ {
		result = timeFrame(ts)
	}
	tcVar = result
}

func BenchmarkDidTry(b *testing.B) {
	sd := newScoreData()
	nick := "Odd"
	channel := "#channel"
	var result bool

	for i := 0; i < b.N; i++ {
		result = sd.didTry(channel, nick)
	}
	boolVar = result
}

