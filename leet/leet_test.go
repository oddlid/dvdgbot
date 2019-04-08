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
	sd := NewScoreData()
	//sd.Get(c)

	fmt.Println("Creating Oddlid")
	o := sd.Get(c).Get("Oddlid")
	fmt.Println("Creating Tord")
	t := sd.Get(c).Get("Tord")
	fmt.Println("Creating Snelhest")
	s := sd.Get(c).Get("Snelhest")
	fmt.Println("Creating bAAAArd")
	b := sd.Get(c).Get("bAAAArd")

	o.Score(10)
	t.Score(8)
	s.Score(6)
	b.Score(4)

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
	n, err := sd.Save(file)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Saved %d bytes to %q", n, fname)
}

func TestBonusPrefixZero(t *testing.T) {
	ts := make([]time.Time, 7)
	ts[0], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.090000000Z")
	ts[1], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:13.370050000Z")
	ts[2], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:01.337000100Z")
	ts[3], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.133700000Z")
	ts[4], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.013370000Z")
	ts[5], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.001337080Z")
	ts[6], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000133700Z")

	exp := 0

	for _, tt := range ts {
		got := bonus(tt)
		ft := fmt.Sprintf("%02d:%02d:%02d.%09d", tt.Hour(), tt.Minute(), tt.Second(), tt.Nanosecond())
		if got != exp {
			t.Errorf("Expected %d, got %d", exp, got)
		}
		t.Logf("Timestamp %q gives %d points bonus, as expected :)", ft, got)
		exp += 1
	}
}

func TestBonusPrefixOther(t *testing.T) {
	ts := make([]time.Time, 7)
	ts[0], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.090000000Z")
	ts[1], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:03.370050000Z")
	ts[2], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:11.337000100Z")
	ts[3], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:20.133700000Z")
	ts[4], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:03.013370000Z")
	ts[5], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.401337080Z")
	ts[6], _ = time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.008133700Z")

	exp := 0

	for _, tt := range ts {
		got := bonus(tt)
		ft := fmt.Sprintf("%02d:%02d:%02d.%09d", tt.Hour(), tt.Minute(), tt.Second(), tt.Nanosecond())
		if got != exp {
			t.Errorf("Expected %d, got %d", exp, got)
		}
		t.Logf("Timestamp %q gives %d points bonus, as expected :)", ft, got)
		//exp += 1
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

func BenchmarkPrefixedBy(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	tstr := fmt.Sprintf("%02d%09d", ts.Second(), ts.Nanosecond())
	sstr := "1337"
	idx := strings.Index(tstr, sstr)
	var bres bool

	for i := 0; i < b.N; i++ {
		bres = pb(tstr, '0', idx)
	}
	boolVar = bres
}

func BenchmarkBonus(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:13.370000000Z")
	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:01.337000000Z")
	//bp := 0
	var ires int
	for i := 0; i < b.N; i++ {
		//bp = bonus(ts)
		ires = bonus(ts)
	}
	intVar = ires
	//b.Logf("bonus: %d", bp)
}

func BenchmarkTryScore(b *testing.B) {
	tm := func(tstr string) time.Time {
		t, _ := time.Parse(time.RFC3339Nano, tstr)
		return t
	}
	sd := NewScoreData()
	channel := "#channel"
	nicks := []struct {
		nick string
		ts time.Time
	}{
		{"Odd_01", tm("2019-04-07T13:37:00.000001337Z")},
		{"Odd_02", tm("2019-04-07T13:37:00.000013370Z")},
		{"Odd_03", tm("2019-04-07T13:37:00.000133700Z")},
	}
	c := sd.Get(channel)
	var bres bool
	var sres string
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, n := range nicks {
			bres, sres = sd.TryScore(channel, n.nick, n.ts)
		}
		//c.MergeScoresForRound(c.GetScoresForRound())
		c.ClearNicksForRound()
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
	sd := NewScoreData()
	nick := "Odd"
	channel := "#channel"
	var result bool

	for i := 0; i < b.N; i++ {
		result = sd.DidTry(channel, nick)
	}
	boolVar = result
}

