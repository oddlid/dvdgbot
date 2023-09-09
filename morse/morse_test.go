package morse

import (
	"testing"
)

func TestAscii2Morse(t *testing.T) {
	str := "SOS"
	morse, err := NewBot().ascii2morse(str)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%q in morse is: %q", str, morse)
}

func TestMorse2Ascii(t *testing.T) {
	str := "... --- ..."
	ascii, err := NewBot().morse2ascii(str)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Morse %q in plain text is: %q", str, ascii)
}
