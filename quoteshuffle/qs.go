package quoteshuffle

/*
This is mostly a new version of oddlid/rndlist with added support for loading/saving via json files.
*/

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	PLUGIN string = "QuoteShuffle"
)

type QuoteData struct {
	FileName string   `json:"-"`
	Src      []string `json:"src"`
	Dst      []string `json:"dst"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func New(fileName string) *QuoteData {
	qd := &QuoteData{FileName: fileName}
	err := qd.LoadSelf()
	if err != nil {
		log.Errorf("%s: New(): Unable to load from %q", PLUGIN, fileName)
	}
	return qd
}

func (qd *QuoteData) Load(r io.Reader) error {
	jb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, qd)
}

func (qd *QuoteData) LoadFile(fileName string) (*QuoteData, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Errorf("%s: LoadFile() Error: %q", PLUGIN, err.Error())
		return qd, err
	}
	defer file.Close()
	err = qd.Load(file)
	if err != nil {
		log.Errorf("%s: LoadFile() Error: %q", PLUGIN, err.Error())
		return qd, err
	}
	log.Debugf("%s: Quotes loaded from file %q", PLUGIN, fileName)
	return qd, err
}

func (qd *QuoteData) LoadSelf() error {
	_, err := qd.LoadFile(qd.FileName)
	return err
}

func (qd *QuoteData) Save(w io.Writer) (int, error) {
	jb, err := json.MarshalIndent(qd, "", "\t")
	//jb, err := json.Marshal(qd)
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func (qd *QuoteData) SaveFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		log.Errorf("%s: SaveFile() Error: %q", PLUGIN, err.Error())
		return err
	}
	defer file.Close()
	n, err := qd.Save(file)
	if err != nil {
		log.Errorf("%s: SaveFile() Error: %q", PLUGIN, err.Error())
		return err
	}
	log.Debugf("%s: Saved %d bytes to %q", PLUGIN, n, fileName)
	return nil
}

func (qd *QuoteData) SaveSelf() error {
	return qd.SaveFile(qd.FileName)
}

func (qd *QuoteData) Add(item string) *QuoteData {
	qd.Src = append(qd.Src, item)
	return qd
}

func (qd *QuoteData) Clear() *QuoteData {
	qd.Src = nil
	qd.Dst = nil
	return qd
}

func (qd *QuoteData) SetList(lst []string) *QuoteData {
	qd.Src = lst
	return qd
}

func (qd *QuoteData) Len() int {
	if nil == qd.Src {
		return -1
	}
	return len(qd.Src)
}

func (qd *QuoteData) Unused() int {
	return qd.Len()
}

func (qd *QuoteData) Used() int {
	if nil == qd.Dst {
		return -1
	}
	return len(qd.Dst)
}

func (qd *QuoteData) rndId() int {
	return rand.Intn(len(qd.Src))
}

func (qd *QuoteData) validId(id int) bool {
	if nil == qd.Src {
		return false
	}
	if id < 0 || id >= qd.Len() {
		return false
	}
	return true
}

func (qd *QuoteData) del(id int) {
	qd.Src[id] = qd.Src[qd.Len()-1]
	//qd.Src[qd.Len()-1] = "" // not really needed?
	qd.Src = qd.Src[:qd.Len()-1]
}

func (qd *QuoteData) Next() string {
	// Restore backing slice if it's been exhausted from previous runs
	if nil == qd.Src || qd.Len() == 0 {
		if nil != qd.Dst && len(qd.Dst) > 0 {
			qd.Src = qd.Dst
			qd.Dst = nil
		} else {
			// If we get here, we don't have any data in either backing slice
			return ""
		}
	}

	id := qd.rndId()
	// this shouldn't really happen, but just to be on the safe side
	if !qd.validId(id) {
		return ""
	}

	item := qd.Src[id] // get a reference to our random item

	// move the item over so it won't get picked again before src list is emptied
	qd.Dst = append(qd.Dst, item)
	qd.del(id)

	return item
}

// Shuffle() is just a wrapper that calls Next() and then SaveSelf()
func (qd *QuoteData) Shuffle() string {
	item := qd.Next()
	qd.SaveSelf()
	return item
}

// Reset() puts whatever is in Dst in Src, and then sets Dst to nil
func (qd *QuoteData) Reset() {
	qd.Src = append(qd.Src, qd.Dst...)
	qd.Dst = nil
}

func (qd *QuoteData) ResetAndSave() {
	if "" == qd.FileName {
		return
	}
	qd.Reset()
	qd.SaveSelf()
}
