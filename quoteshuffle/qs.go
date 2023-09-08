package quoteshuffle

/*
This is mostly a new version of oddlid/rndlist with added support for loading/saving via json files.
*/

import (
	"encoding/json"
	"io"
	"math/rand"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	plugin string = "QuoteShuffle"
)

type QuoteData struct {
	zlog     zerolog.Logger
	FileName string   `json:"-"`
	Src      []string `json:"src"`
	Dst      []string `json:"dst"`
}

func New(fileName string) (*QuoteData, error) {
	return (&QuoteData{
		FileName: fileName,
		zlog:     log.With().Str("plugin", plugin).Logger(),
	}).loadSelf()
}

func (qd *QuoteData) load(r io.Reader) error {
	jb, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, qd)
}

func (qd *QuoteData) loadFile(fileName string) (*QuoteData, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return qd, err
	}
	defer file.Close()

	if err = qd.load(file); err != nil {
		return qd, err
	}
	qd.zlog.Debug().
		Str("filename", fileName).
		Msg("Quotes loaded from file")

	return qd, nil
}

func (qd *QuoteData) loadSelf() (*QuoteData, error) {
	return qd.loadFile(qd.FileName)
}

func (qd *QuoteData) save(w io.Writer) (int, error) {
	jb, err := json.MarshalIndent(qd, "", "\t")
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func (qd *QuoteData) saveFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := qd.save(file)
	if err != nil {
		return err
	}
	qd.zlog.Debug().
		Str("filename", fileName).
		Int("bytes", n).
		Msg("File saved")
	return nil
}

func (qd *QuoteData) saveSelf() error {
	return qd.saveFile(qd.FileName)
}

func (qd *QuoteData) len() int {
	if nil == qd.Src {
		return -1
	}
	return len(qd.Src)
}

func (qd *QuoteData) rndID() int {
	//nolint:gosec // sufficient
	return rand.Intn(len(qd.Src))
}

func (qd *QuoteData) validID(id int) bool {
	if qd.Src == nil {
		return false
	}
	if id < 0 || id >= qd.len() {
		return false
	}
	return true
}

func (qd *QuoteData) del(id int) {
	qd.Src[id] = qd.Src[qd.len()-1]
	// qd.Src[qd.Len()-1] = "" // not really needed?
	qd.Src = qd.Src[:qd.len()-1]
}

func (qd *QuoteData) next() string {
	// Restore backing slice if it's been exhausted from previous runs
	if qd.len() == 0 {
		if len(qd.Dst) > 0 {
			qd.Src = qd.Dst
			qd.Dst = nil
		} else {
			// If we get here, we don't have any data in either backing slice
			return ""
		}
	}

	id := qd.rndID()
	// this shouldn't really happen, but just to be on the safe side
	if !qd.validID(id) {
		return ""
	}

	item := qd.Src[id] // get a reference to our random item

	// move the item over so it won't get picked again before src list is emptied
	qd.Dst = append(qd.Dst, item)
	qd.del(id)

	return item
}

// Shuffle() is just a wrapper that calls next() and then saveSelf()
func (qd *QuoteData) Shuffle() (string, error) {
	return qd.next(), qd.saveSelf()
}
