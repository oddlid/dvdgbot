package leet

import (
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type User struct {
	sync.RWMutex
	Nick      string    `json:"nick"`       // duplicate of map key, but we need to have it here as well sometimes
	Points    int       `json:"score"`      // current points total
	LastEntry time.Time `json:"last_entry"` // time of last !1337 post that resulted in a score, positive or negative
	Locked    bool      `json:"locked"`     // true if the user has reached the target limit
	didTry    bool
	l         *logrus.Entry
}

type UserMap map[string]*User
type UserSlice []*User

//func (um UserMap) toSlice() UserSlice {
//	us := make(UserSlice, len(um))
//	i := 0
//	for _, v := range um {
//		us[i] = v
//		i++
//	}
//	return us
//}

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

//func (um UserMap) splitByPointLimit(limit int) (below, at, above UserSlice) {
//	maxLen := len(um)
//	below = make(UserSlice, 0, maxLen)
//	at = make(UserSlice, 0, maxLen)
//	above = make(UserSlice, 0, maxLen)
//
//	for _, v := range um {
//		points := v.getScore()
//		if points < limit {
//			below = append(below, v)
//		} else if points == limit {
//			at = append(at, v)
//		} else if points > limit {
//			above = append(above, v)
//		}
//	}
//	return
//}

func (us UserSlice) sortByLastEntryAsc() UserSlice {
	sort.Slice(us,
		func(i, j int) bool {
			return us[i].LastEntry.Before(us[j].LastEntry)
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
func (u *User) score(points int) (bool, int) {
	if u.hasTried() {
		return false, u.getScore()
	}
	u.try(true)
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

func (u *User) getShortTime() string {
	// use 0's to get subsecond value padded,
	// use 9's to get trailing 0's removed.
	// I don't know yet why, but when running tests on my mac,
	// I always get the last 3 digits as 0 when using padding,
	// although they're never 0 when calling t.Nanoseconds()
	// other places in the code.
	// TODO: Try to get full precision here
	return u.getLastEntry().Format("15:04:05.000000000")
}

func (u *User) getLongDate() string {
	return u.getLastEntry().Format("2006-01-02 15:04:05.999999999")
}

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
