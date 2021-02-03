package leet

// KV is just used for sorting nicks and corresponding scores,
// and printing them aligned.
type KV struct {
	Key string
	Val int
}

type KVList []KV

func (kv KV) KeyLen() int {
	return len(kv.Key)
}

// Len returns the number of items in the list. To satisfy sort.Interface
func (kvl KVList) Len() int {
	return len(kvl)
}

// Less returns true if the first KV has a lower values than the second KV.
// To satisfy sort.Interface
func (kvl KVList) Less(i, j int) bool {
	return kvl[i].Val < kvl[j].Val
}

// Satisfy sort.Interface
func (kvl KVList) Swap(i, j int) {
	kvl[i], kvl[j] = kvl[j], kvl[i]
}

// LongestKey returns the length of the longest key in the slice.
// Used for aligning when printing.
func (kvl KVList) LongestKey() int {
	key_maxlen := 0
	for _, kv := range kvl {
		klen := kv.KeyLen()
		if klen > key_maxlen {
			key_maxlen = klen
		}
	}
	return key_maxlen
}
