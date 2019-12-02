package quoteshuffle

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

const (
	QFILE string = "quotes.json"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestShuffle(t *testing.T) {
	qd := New(QFILE)

	for i := 0; i < 3; i++ {
		t.Logf("Quote: %q", qd.Shuffle())
	}
}

// This is mostly for resetting the quotes file via the go test command easily
func TestReset(t *testing.T) {
	qd := New(QFILE)
	t.Logf("%s: %d unused, and %d used quotes", QFILE, qd.Unused(), qd.Used())
	qd.ResetAndSave()
	t.Logf("%s: %d unused, and %d used quotes", QFILE, qd.Unused(), qd.Used())
}
