package quoteshuffle

import (
	"testing"

	log "github.com/Sirupsen/logrus"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestShuffle(t *testing.T) {
	qd := New("quotes.json")

	for i := 0; i < 3; i++ {
		t.Logf("Quote: %q", qd.Shuffle())
	}
}
