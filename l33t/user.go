package l33t

import "sync/atomic"

// User is a transient wrapper for UserData
type User struct {
	data          *UserData
	name          string
	activeInRound atomic.Bool
}
