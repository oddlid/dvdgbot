package quoteshuffle

import (
	"testing"

	"github.com/rs/zerolog"
)

const (
	QFILE string = "quotes.json"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func TestShuffle(t *testing.T) {
	qd, err := New(QFILE)
	if nil != err {
		t.Error(err)
	}

	for i := 0; i < 3; i++ {
		t.Logf("Quote: %q", qd.Shuffle())
	}
}

// This is mostly for resetting the quotes file via the go test command easily
func TestReset(t *testing.T) {
	qd, err := New(QFILE)
	if nil != err {
		t.Error(err)
	}
	t.Logf("%s: %d unused, and %d used quotes", QFILE, qd.Unused(), qd.Used())
	qd.ResetAndSave()
	t.Logf("%s: %d unused, and %d used quotes", QFILE, qd.Unused(), qd.Used())
}
