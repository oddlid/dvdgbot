package l33t

type UserData struct {
	Entry   EntryTime    `json:"entry"`
	Taxes   ValueTracker `json:"taxes"`   // how much tax over time
	Bonuses ValueTracker `json:"bonuses"` // how much bonuses over time
	Misses  ValueTracker `json:"misses"`  // how many times have the user been early or late
	Points  int          `json:"score"`   // current points total
	Done    bool         `json:"done"`    // true if the user has reached the target limit
}

type UserDataMap map[string]*UserData

// getUser returns a User for the data associated with the nick, if it exists,
// otherwise it creates and inserts an empty data entry for the given given nick
// and returns a new User for that
func (udm UserDataMap) getUser(nick string) *User {
	if udm == nil {
		return nil
	}
	userData, ok := udm[nick]
	if ok {
		return &User{
			name: nick,
			data: userData,
		}
	}

	udm[nick] = &UserData{}
	return udm.getUser(nick)
}
