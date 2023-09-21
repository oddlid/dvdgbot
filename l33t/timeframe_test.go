package l33t

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_TimeCode_insideWindow(t *testing.T) {
	t.Parallel()

	assert.False(t, tcBefore.insideWindow())
	assert.True(t, tcEarly.insideWindow())
	assert.True(t, tcOnTime.insideWindow())
	assert.True(t, tcLate.insideWindow())
	assert.False(t, tcAfter.insideWindow())
}

func Test_TimeCode_nearMiss(t *testing.T) {
	t.Parallel()

	assert.False(t, tcBefore.nearMiss())
	assert.True(t, tcEarly.nearMiss())
	assert.False(t, tcOnTime.nearMiss())
	assert.True(t, tcLate.nearMiss())
	assert.False(t, tcAfter.nearMiss())
}

func Test_TimeFrame_getCronTime(t *testing.T) {
	t.Parallel()

	tf := TimeFrame{
		hour:         13,
		minute:       37,
		windowBefore: time.Minute,
		windowAfter:  time.Minute,
	}
	now := time.Now()
	adjust := time.Minute
	cronTime := tf.getCronTime(now, adjust)
	assert.Equal(t, tf.hour, cronTime.hour)
	assert.Equal(t, tf.minute+1, cronTime.minute)
}

func Test_TimeFrame_asCronSpec(t *testing.T) {
	t.Parallel()

	tf := TimeFrame{
		hour:         13,
		minute:       37,
		windowBefore: time.Minute,
		windowAfter:  time.Minute,
	}

	assert.Equal(t, "13 37 * * *", tf.asCronSpec())
}

func Test_TimeFrame_code(t *testing.T) {
	t.Parallel()

	now := time.Date(
		2023,
		time.September,
		13,
		13,
		37,
		0,
		0,
		time.UTC,
	)
	tf := TimeFrame{
		hour:         now.Hour(),
		minute:       now.Minute(),
		windowBefore: time.Minute,
		windowAfter:  time.Minute,
	}
	assert.Equal(t, tcOnTime, tf.code(now))

	tf.hour += 2
	assert.Equal(t, tcBefore, tf.code(now))

	tf.hour -= 4
	assert.Equal(t, tcAfter, tf.code(now))

	tf.hour = now.Hour()
	tf.minute = now.Minute() + 2
	assert.Equal(t, tcBefore, tf.code(now))

	tf.minute = now.Minute() - 2
	assert.Equal(t, tcAfter, tf.code(now))

	tf.minute = now.Minute() + 1
	assert.Equal(t, tcEarly, tf.code(now))

	tf.minute = now.Minute() - 1
	assert.Equal(t, tcLate, tf.code(now))
}

func Test_TimeFrame_getTargetScore(t *testing.T) {
	t.Parallel()

	tf := TimeFrame{hour: 13, minute: 37}
	assert.Equal(t, 1337, tf.getTargetScore())
}

func Test_TimeFrame_distance(t *testing.T) {
	t.Parallel()

	t1 := time.Date(
		0,
		1,
		1,
		13,
		37,
		0,
		0,
		time.UTC,
	)
	tf := TimeFrame{
		hour:   t1.Hour(),
		minute: t1.Minute(),
	}

	t2 := t1.Add(1 * time.Nanosecond)
	assert.Equal(t, time.Nanosecond, tf.distance(t2))

	t2 = t1.Add(-1 * time.Nanosecond)
	assert.Equal(t, time.Nanosecond, tf.distance(t2))

	t2 = t1.Add(5 * time.Millisecond)
	assert.Equal(t, 5*time.Millisecond, tf.distance(t2))

	t2 = t1.Add(-5 * time.Millisecond)
	assert.Equal(t, 5*time.Millisecond, tf.distance(t2))
}
