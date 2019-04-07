package leet

import (
	"fmt"
	"os"
	"testing"
	"time"
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

func BenchmarkBonus(b *testing.B) {
	ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:00.000001337Z")
	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:13.370000000Z")
	//ts, _ := time.Parse(time.RFC3339Nano, "2019-04-07T13:37:01.337000000Z")
	bp := 0
	for i := 0; i < b.N; i++ {
		bp = bonus(ts)
	}
	b.Logf("bonus: %d", bp)
}


