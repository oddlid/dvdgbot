package l33t

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValueTracker_add(t *testing.T) {
	t.Parallel()

	assert.NotPanics(
		t,
		func() {
			(*ValueTracker)(nil).add(1)
		},
	)

	vt := ValueTracker{}
	vt.add(2)
	assert.Equal(t, 2, vt.Total)
	assert.Equal(t, 1, vt.Times)
}
