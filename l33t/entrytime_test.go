package l33t

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_EntryTime_update(t *testing.T) {
	t.Parallel()

	assert.NotPanics(
		t,
		func() {
			(*EntryTime)(nil).update(TimeFrame{}, time.Time{})
		},
	)

	tf := TimeFrame{
		hour:         13,
		minute:       37,
		windowBefore: time.Minute,
		windowAfter:  time.Minute,
	}
	et := EntryTime{}

	tt := time.Date(2023, 9, 16, 18, 39, 45, 0, time.UTC)
	et.update(tf, tt)
	assert.True(t, et.Last.IsZero())
	assert.True(t, et.Best.IsZero())

	tt = time.Date(2023, 9, 16, tf.hour, tf.minute-1, 59, 0, time.UTC)
	et.update(tf, tt)
	assert.True(t, tt.Equal(et.Last))
	assert.True(t, tt.Equal(et.Best))

	et.update(tf, tt)
	assert.True(t, tt.Equal(et.Best))

	tt = time.Date(2023, 9, 16, tf.hour, tf.minute, 0, 3, time.UTC)
	et.update(tf, tt)
	assert.True(t, tt.Equal(et.Best))

	tt = time.Date(2023, 9, 16, tf.hour, tf.minute, 0, 2, time.UTC)
	et.update(tf, tt)
	assert.True(t, tt.Equal(et.Best))
}
