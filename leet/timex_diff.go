package leet

// This file is copied and modified from:
// https://github.com/icza/gox/blob/master/timex/timex.go

import (
	"fmt"
	"strings"
	"time"
)

type timex struct {
	year   int
	month  int
	day    int
	hour   int
	minute int
	second int
}

// Diff calculates the absolute difference between 2 time instances in
// years, months, days, hours, minutes and seconds.
//
// For details, see https://stackoverflow.com/a/36531443/1705598
func timexDiff(a, b time.Time) timex {
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

	tx := timex{
		year:   y2 - y1,
		month:  int(M2 - M1),
		day:    d2 - d1,
		hour:   h2 - h1,
		minute: m2 - m1,
		second: s2 - s1,
	}

	// Normalize negative values
	if tx.second < 0 {
		tx.second += 60
		tx.minute--
	}
	if tx.minute < 0 {
		tx.minute += 60
		tx.hour--
	}
	if tx.hour < 0 {
		tx.hour += 24
		tx.day--
	}
	if tx.day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		tx.day += 32 - t.Day()
		tx.month--
	}
	if tx.month < 0 {
		tx.month += 12
		tx.year--
	}

	return tx
}

// timexString is just a helper func I added myself for getting dynamic output from the diff above
func (tx timex) String() string {
	var sb strings.Builder

	plural := func(word string, amount int, addComma bool) {
		if amount == 1 {
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
	if tx.year > 0 {
		plural("year", tx.year, true)
		plural("month", tx.month, true)
		plural("day", tx.day, true)
		plural("hour", tx.hour, true)
		plural("minute", tx.minute, true)
		plural("second", tx.second, false)
	} else {
		if tx.month > 0 {
			plural("month", tx.month, true)
			plural("day", tx.day, true)
			plural("hour", tx.hour, true)
			plural("minute", tx.minute, true)
			plural("second", tx.second, false)
		} else {
			if tx.day > 0 {
				plural("day", tx.day, true)
				plural("hour", tx.hour, true)
				plural("minute", tx.minute, true)
				plural("second", tx.second, false)
			} else {
				if tx.hour > 0 {
					plural("hour", tx.hour, true)
					plural("minute", tx.minute, true)
					plural("second", tx.second, false)
				} else {
					if tx.minute > 0 {
						plural("minute", tx.minute, true)
						plural("second", tx.second, false)
					} else {
						plural("second", tx.second, false)
					}
				}
			}
		}
	}

	return sb.String()
}
