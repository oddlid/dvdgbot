package leet

import (
	"testing"
)

func Test_TimeFrame_getTargetScore(t *testing.T) {
	t.Parallel()

	tf := TimeFrame{hour: 13, minute: 37}
	got := tf.getTargetScore()
	if got != 1337 {
		t.Fatalf("Expected 1337, got: %d", got)
	}
}
