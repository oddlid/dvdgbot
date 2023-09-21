package l33t

type Channel struct {
	name        string
	data        *ChannelData
	contestants []*User // temp storage for each round
}
