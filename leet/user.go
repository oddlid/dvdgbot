package leet

import (
	"sync"
	"time"
)

type User struct {
	sync.RWMutex
	Points int `json:"score"`
	didTry bool
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
