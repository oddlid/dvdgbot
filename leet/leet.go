package leet

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-chat-bot/bot"
)

var (
	isFirst  bool = true
	botstart time.Time
	scores   map[string]int
)

type KV struct {
	Key string
	Val int
}

type KVList []KV

func rank() KVList {
	kl := make(KVList, len(scores))
	i := 0
	for k, v := range scores {
		kl[i] = KV{k, v}
		i++
	}
	sort.Sort(sort.Reverse(kl))
	return kl
}

func (kl KVList) Len() int {
	return len(kl)
}

func (kl KVList) Less(i, j int) bool {
	return kl[i].Val < kl[j].Val
}

func (kl KVList) Swap(i, j int) {
	kl[i], kl[j] = kl[j], kl[i]
}

func leet(cmd *bot.Cmd) (string, error) {
	//user := cmd.User.Nick
	if len(cmd.Args) == 2 && cmd.Args[1] == "stats" {
		kl := rank()
		str := fmt.Sprintf("Stats since %s:\n", botstart)
		for _, kv := range kl {
			str += fmt.Sprintf("%s: %d\n", kv.Key, kv.Val)
		}
		return str, nil
	}

	//	t := time.Now()
	//	if t.Hour() == 13 && t.Minute() == 37 {
	//	}

	if isFirst {
		scores[cmd.User.RealName] += 2
		isFirst = false
		time.AfterFunc(1*time.Minute, func() {
			isFirst = true
		})
	} else {
		scores[cmd.User.RealName]++
	}
	return fmt.Sprintf("Whoop! %s total score: %d\n", cmd.User.RealName, scores[cmd.User.RealName]), nil
}

func init() {
	botstart = time.Now()
	scores = make(map[string]int)

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"",
		leet)
}
