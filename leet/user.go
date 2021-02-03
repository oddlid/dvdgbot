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
	didTry    bool
	l         *logrus.Entry
}

type UserMap map[string]*User
type UserSlice []*User

func (um UserMap) toSlice() UserSlice {
	us := make(UserSlice, len(um))
	i := 0
	for _, v := range um {
		us[i] = v
		i++
	}
	return us
}

func (um UserMap) filterByPointsEQ(points int) UserSlice {
	us := make(UserSlice, 0, len(um))
	for _, v := range um {
		if points == v.Points {
			us = append(us, v)
		}
	}
	return us
}

func (us UserSlice) sortByLastEntryAsc() {
	sort.Slice(us,
		func(i, j int) bool {
			return us[i].LastEntry.Before(us[j].LastEntry)
		},
	)
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

// wrapper around score()
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
