package leet

// This file is copied and modified from:
// https://github.com/icza/gox/blob/master/timex/timex.go

import (
	"time"
	"fmt"
	"strings"
)

// Diff calculates the absolute difference between 2 time instances in
// years, months, days, hours, minutes and seconds.
//
// For details, see https://stackoverflow.com/a/36531443/1705598
func timexDiff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}

// timexString is just a helper func I added myself for getting dynamic output from the diff above
func timexString(year, month, day, hour, min, sec int) string {
	var sb strings.Builder

	plural := func(word string, amount int, addComma bool) {
		if 1 == amount {
			sb.WriteString(fmt.Sprintf("%d %s", amount, word))
		} else {
			sb.WriteString(fmt.Sprintf("%d %ss", amount, word))
		}
		if addComma {
			sb.WriteString(", ")
		}
	}

	// We want to skip prepending output for values that are 0,
	// but only if the value before it was also 0
	if year > 0 {
		plural("year", year, true)
		plural("month", month, true)
		plural("day", day, true)
		plural("hour", hour, true)
		plural("minute", min, true)
		plural("second", sec, false)
	} else {
		if month > 0 {
			plural("month", month, true)
			plural("day", day, true)
			plural("hour", hour, true)
			plural("minute", min, true)
			plural("second", sec, false)
		} else {
			if day > 0 {
				plural("day", day, true)
				plural("hour", hour, true)
				plural("minute", min, true)
				plural("second", sec, false)
			} else {
				if hour > 0 {
					plural("hour", hour, true)
					plural("minute", min, true)
					plural("second", sec, false)
				} else {
					if min > 0 {
						plural("minute", min, true)
						plural("second", sec, false)
					} else {
						plural("second", sec, false)
					}
				}
			}
		}
	}

	return sb.String()
}
