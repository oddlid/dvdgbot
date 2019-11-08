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

func (u *User) AddScore(points int) int {
	u.Lock()
	defer u.Unlock()
	u.Points += points
	return u.Points
}

func (u *User) Score(points int) (bool, int) {
	if u.didTry {
		return false, u.Points
	}
	u.AddScore(points)
	u.Lock()
	u.didTry = true
	total := u.Points
	u.Unlock()
	// Reset didTry after 2 minutes
	// This should create a "loophole" so that if a user posts too early and gets -1,
	// they could manage to get another -1 by being too late as well :D
	time.AfterFunc(2*time.Minute, func() {
		u.Lock()
		u.didTry = false
		u.Unlock()
	})
	return true, total
}

