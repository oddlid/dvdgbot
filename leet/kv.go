package leet

type KV struct {
	Key string
	Val int
}

type KVList []KV

func (kl KVList) Len() int {
	return len(kl)
}

func (kl KVList) Less(i, j int) bool {
	return kl[i].Val < kl[j].Val
}

func (kl KVList) Swap(i, j int) {
	kl[i], kl[j] = kl[j], kl[i]
}

