package leet

import (
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)


type ScoreTracker struct {
	Times int `json:"times"` // how many times have the user gotten a bonus or tax
	Total int `json:"total"` // the sum of all
}

type User struct {
	sync.RWMutex
	Nick      string       `json:"nick"`       // duplicate of map key, but we need to have it here as well sometimes
	Points    int          `json:"score"`      // current points total
	Locked    bool         `json:"locked"`     // true if the user has reached the target limit
	LastEntry time.Time    `json:"last_entry"` // time of last !1337 post that resulted in a score, positive or negative
	BestEntry time.Time    `json:"best_entry"` // tighhtest to 1337, or whatever...
	Taxes     ScoreTracker `json:"taxes"`      // hos much tax over time
	Bonuses   ScoreTracker `json:"bonuses"`    // how much bonuses over time
	didTry    bool
	l         *logrus.Entry
}

type UserMap map[string]*User
type UserSlice []*User

func (um UserMap) toSlice() UserSlice {
	us := make(UserSlice, 0, len(um))
	for _, v := range um {
		us = append(us, v)
	}
	return us
}

func (um UserMap) filterByPointsEQ(points int) UserSlice {
	us := make(UserSlice, 0, len(um))
	for _, v := range um {
		if v.getScore() == points {
			us = append(us, v)
		}
	}
	return us
}

func (um UserMap) filterByLocked(locked bool) UserSlice {
	us := make(UserSlice, 0, len(um))
	for _, v := range um {
		if locked == v.Locked {
			us = append(us, v)
		}
	}
	return us
}

func (um UserMap) longestNickLen() int {
	maxlen := 0
	for k := range um {
		nlen := len(k)
		if nlen > maxlen {
			maxlen = nlen
		}
	}
	return maxlen
}

func (us UserSlice) sortByLastEntryAsc() UserSlice {
	sort.Slice(us,
		func(i, j int) bool {
			return us[i].LastEntry.Before(us[j].LastEntry)
		},
	)
	return us
}

func (us UserSlice) sortByPointsDesc() UserSlice {
	sort.Slice(us,
		func(i, j int) bool {
			return us[i].Points > us[j].Points
		},
	)
	return us
}

// use this to get "rank" after sorting by date
func (us UserSlice) getIndex(nick string) int {
	for idx, u := range us {
		if nick == u.Nick {
			return idx
		}
	}
	return -1
}

func (u *User) log() *logrus.Entry {
	if nil == u.l {
		return _log // pkg global
	}
	return u.l
}

func (u *User) try(val bool) {
	u.Lock()
	u.didTry = val
	u.Unlock()
}

func (u *User) hasTried() bool {
	u.RLock()
	defer u.RUnlock()
	return u.didTry
}

func (u *User) getScore() int {
	u.RLock()
	defer u.RUnlock()
	return u.Points
}

func (u *User) addScore(points int) int {
	u.Lock()
	defer u.Unlock()
	u.Points += points
	return u.Points
}

// mostly for testing at the time of writing
func (u *User) setScore(points int) {
	u.Lock()
	u.Points = points
	u.Unlock()
}

// wrapper around addScore()
func (u *User) score(points int, when time.Time) (bool, int) {
	if u.hasTried() {
		return false, u.getScore()
	}
	u.try(true)
	u.setLastEntry(when)
	go u.setBestEntry(when) // run in goroutine in order to not take time from others scoring
	// Reset didTry after 2 minutes
	// This should create a "loophole" so that if a user posts too early and gets -1,
	// they could manage to get another -1 by being too late as well :D
	time.AfterFunc(2*time.Minute, func() {
		u.try(false)
	})
	return true, u.addScore(points)
}

func (u *User) getLastEntry() time.Time {
	u.RLock()
	defer u.RUnlock()
	return u.LastEntry
}

func (u *User) setLastEntry(when time.Time) {
	u.Lock()
	u.LastEntry = when
	u.Unlock()
}

//func (u *User) getShortTime() string {
//	// use 0's to get subsecond value padded,
//	// use 9's to get trailing 0's removed.
//	// I don't know yet why, but when running tests on my mac,
//	// I always get the last 3 digits as 0 when using padding,
//	// although they're never 0 when calling t.Nanoseconds()
//	// other places in the code.
//	// TODO: Try to get full precision here
//	return u.getLastEntry().Format("15:04:05.000000000")
//}

//func (u *User) getLongDate() string {
//	return u.getLastEntry().Format("2006-01-02 15:04:05.000000000")
//}

func (u *User) setLocked(locked bool) {
	u.Lock()
	u.Locked = locked
	u.Unlock()
}

func (u *User) isLocked() bool {
	u.RLock()
	defer u.RUnlock()
	return u.Locked
}

// Helper functions for comparing times. We can't use time.(Defore|After), since
// we only want to compare the time part, not the date
func IsBefore(t1, t2 time.Time) bool {
	if t1.Minute() < t2.Minute() {
		return true
	}
	if t1.Minute() == t2.Minute() {
		if t1.Second() < t2.Second() {
			return true
		}
		if t1.Second() == t2.Second() {
			if t1.Nanosecond() < t2.Nanosecond() {
				return true
			}
		}
	}
	return false
}

func IsAfter(t1, t2 time.Time) bool {
	if t1.Minute() > t2.Minute() {
		return true
	}
	if t1.Minute() == t2.Minute() {
		if t1.Second() > t2.Second() {
			return true
		}
		if t1.Second() == t2.Second() {
			if t1.Nanosecond() > t2.Nanosecond() {
				return true
			}
		}
	}
	return false
}

func (u *User) getBestEntry() time.Time {
	u.RLock()
	defer u.RUnlock()
	return u.BestEntry
}

func (u *User) setBestEntryWithLock(when time.Time) {
	u.Lock()
	u.BestEntry = when
	u.Unlock()
}

// setBestEntry() will set BestEntry for the user, if given time is closer to target
// time than previously stored time value
func (u *User) setBestEntry(when time.Time) {
	llog := u.log().WithFields(logrus.Fields{
		"func":     "setBestEntry",
		"oldEntry": u.BestEntry,
		"newEntry": when,
	})
	// If no previous value, we just don't care and set what we get
	if u.BestEntry.IsZero() {
		llog.Debug("No previous value, accepting anything")
		u.setBestEntryWithLock(when)
		return
	}
	// ...
	within, newTimeCode := withinTimeFrame(when)
	if !within {
		llog.WithField("newTimeCode", newTimeCode).Debug("Outside timeframe")
		return
	}

	// If we're here, it means we're within +-1 minute of the target (13:37)

	// Note to self:
	// We don't need to compare a time that is late to a time that is early, as the early
	// time will (almost) always be closer to target than a late time, since a time after will
	// always be at least a minute off, while a time before is at most a minute off.
	// That is:
	// Late = at least 60+ seconds after
	// Early = at most 59- seconds before

	// We don't need to check newTimeCode for TF_BEFORE or TF_AFTER, since that would be outside
	// timeframe, and so the check above returns if that's the case.
	// We still check oldTimeCode for every variant though, as it could have been set to anything
	// the first time this func is called, when the previous value is empty.

	oldTimeCode := timeFrame(u.BestEntry)

	if TF_BEFORE == oldTimeCode || TF_EARLY == oldTimeCode {
		if TF_EARLY == newTimeCode {
			if IsAfter(when, u.BestEntry) {
				llog.Debug("Both times are before, but new time is better - setting time")
				u.setBestEntryWithLock(when)
				return
			}
			llog.Debug("Both times are before, but old one is better - skipping")
			return
		}
		if TF_ONTIME == newTimeCode {
			llog.Debug("Old time is before, new time is on time - setting time")
			u.setBestEntryWithLock(when)
			return
		}
		if TF_LATE == newTimeCode {
			llog.Debug("Old time is before, new time is after - skipping")
			return
		}
		llog.Debug("Old time is before, new time unchecked")
		return
	}


	if TF_ONTIME == oldTimeCode {
		if TF_EARLY == newTimeCode {
			llog.Debug("Old time on time, but new is before - skipping")
			return
		}
		if TF_LATE == newTimeCode {
			llog.Debug("Old time on time, but new time after - skipping")
			return
		}
		if IsBefore(when, u.BestEntry) {
			llog.Debug("Both times on time, but new one is better - setting time")
			u.setBestEntryWithLock(when)
			return
		}
		llog.Debug("Both times on time, but old one is better - skipping")
		return
	}

	if (TF_LATE == oldTimeCode || TF_AFTER == oldTimeCode) {
		if TF_EARLY == newTimeCode {
			// Most likely, an early time will be closer to the target than a
			// late time
			llog.Debug("Old time is after, new time before - setting time")
			u.setBestEntryWithLock(when)
			return
		}
		if TF_ONTIME == newTimeCode {
			llog.Debug("Old time is after, new time is on time - setting time")
			u.setBestEntryWithLock(when)
			return
		}
		if IsBefore(when, u.BestEntry) {
			llog.Debug("Both times are after, but new time is better - setting time")
			u.setBestEntryWithLock(when)
			return
		}
		llog.Debug("Both times are after, but old one is better - skipping")
		return
	}

	llog.Debug("Should not get here")
}

func (u *User) getTaxTotal() int {
	u.RLock()
	defer u.RUnlock()
	return u.Taxes.Total
}

func (u *User) getTaxTimes() int {
	u.RLock()
	defer u.RUnlock()
	return u.Taxes.Times
}

func (u *User) addTax(tax int) {
	u.Lock()
	u.Taxes.Total += tax
	if tax > 0 { // we don't want the counter to step up if tax is 0
		u.Taxes.Times++
	}
	u.Unlock()
}

func (u *User) getBonusTotal() int {
	u.RLock()
	defer u.RUnlock()
	return u.Bonuses.Total
}

func (u *User) getBonusTimes() int {
	u.RLock()
	defer u.RUnlock()
	return u.Bonuses.Times
}

func (u *User) addBonus(bonus int) {
	u.Lock()
	u.Bonuses.Total += bonus
	if bonus > 0 { // we don't want the counter to step up if bonus is 0
		u.Bonuses.Times++
	}
	u.Unlock()
}
