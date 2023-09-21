package l33t

import "time"

type EntryTime struct {
	Last time.Time `json:"last_entry"`
	Best time.Time `json:"best_entry"`
}

func (et *EntryTime) update(tf TimeFrame, t time.Time) {
	if et == nil {
		return
	}

	tcForThisEntry := tf.code(t)

	// Don't update anything if outside time window
	if !tcForThisEntry.insideWindow() {
		return
	}

	et.Last = t

	// First time, nothing to compare to
	if et.Best.IsZero() {
		et.Best = t
		return
	}

	// We don't want to overwrite the previous best with this one, if this one
	// is closer, but a near miss
	if tcForThisEntry.nearMiss() {
		return
	}

	// If the previous best was set regardless because it was the first entry, it might
	// be a near miss, and if so, we know that this entry is better without comparing more
	if tf.code(et.Best).nearMiss() {
		et.Best = t
		return
	}

	// Now we know both the previous best and this entry is on time, so we must compare
	// which is closest, and set the best
	if tf.distance(t) < tf.distance(et.Best) {
		et.Best = t
	}
}
