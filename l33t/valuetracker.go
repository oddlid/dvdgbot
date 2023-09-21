package l33t

type ValueTracker struct {
	Times int `json:"times"`
	Total int `json:"total"`
}

func (vt *ValueTracker) add(value int) {
	if vt == nil {
		return
	}
	vt.Total += value
	if value != 0 {
		vt.Times++
	}
}
