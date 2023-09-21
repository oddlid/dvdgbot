package l33t

import (
	"fmt"
	"time"
)

type TimeCode uint8

// Constants for signaling offset from time window
const (
	tcBefore TimeCode = iota // more than a minute before
	tcEarly                  // less than a minute before
	tcOnTime                 // within correct minute
	tcLate                   // less than a minute late
	tcAfter                  // more than a minute late
)

type TimeFrame struct {
	hour         int
	minute       int
	windowBefore time.Duration
	windowAfter  time.Duration
}

func (tc TimeCode) String() string {
	switch tc {
	case tcBefore:
		return "before"
	case tcEarly:
		return "early"
	case tcLate:
		return "late"
	case tcAfter:
		return "after"
	default:
		return "on time"
	}
}

func (tc TimeCode) insideWindow() bool {
	return tc == tcEarly || tc == tcOnTime || tc == tcLate
}

func (tc TimeCode) nearMiss() bool {
	return tc == tcEarly || tc == tcLate
}

func (tf TimeFrame) getCronTime(t time.Time, adjust time.Duration) TimeFrame {
	then := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		tf.hour,
		tf.minute,
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	when := then.Add(adjust)
	return TimeFrame{
		hour:         when.Hour(),
		minute:       when.Minute(),
		windowBefore: tf.windowBefore,
		windowAfter:  tf.windowAfter,
	}
}

func (tf TimeFrame) asCronSpec() string {
	return fmt.Sprintf("%d %d * * *", tf.hour, tf.minute)
}

func (tf TimeFrame) code(t time.Time) TimeCode {
	switch h := t.Hour(); {
	case h < tf.hour:
		return tcBefore
	case h > tf.hour:
		return tcAfter
	}

	switch m := t.Minute(); {
	case m < tf.minute-int(tf.windowBefore.Minutes()):
		return tcBefore
	case m > tf.minute+int(tf.windowAfter.Minutes()):
		return tcAfter
	case m == tf.minute-int(tf.windowBefore.Minutes()):
		return tcEarly
	case m == tf.minute+int(tf.windowAfter.Minutes()):
		return tcLate
	default:
		return tcOnTime
	}
}

func (tf TimeFrame) getTargetScore() int {
	return tf.hour*100 + tf.minute
}

func (tf TimeFrame) distance(t time.Time) time.Duration {
	t2 := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		tf.hour,
		tf.minute,
		0,
		0,
		t.Location(),
	)
	if t.After(t2) {
		return t.Sub(t2)
	}
	return t2.Sub(t)
}
