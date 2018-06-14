package leet

import (
	"testing"
	"os"
	"fmt"
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
