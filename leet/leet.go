package leet

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-chat-bot/bot"
)

const (
	DEF_HOUR   int = 13
	DEF_MINUTE int = 37
)

var (
	hour     int  = DEF_HOUR
	minute   int  = DEF_MINUTE
	isFirst  bool = true
	botstart time.Time
	scores   map[string]int
	didTry   map[string]bool
)

type KV struct {
	Key string
	Val int
}

type KVList []KV

func rank() (KVList, int) {
	kl := make(KVList, len(scores))
	i := 0
	nick_maxlen := 0
	for k, v := range scores {
		kl[i] = KV{k, v}
		i++
		nlen := len(k)
		if nlen > nick_maxlen {
			nick_maxlen = nlen
		}
	}
	sort.Sort(sort.Reverse(kl))
	return kl, nick_maxlen
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

// tempBlockUser prevents a nick from entering more than once in 2 minutes
func tempBlockUser(nick string) {
	didTry[nick] = true
	time.AfterFunc(2*time.Minute, func() {
		didTry[nick] = false
	})
}

func leet(cmd *bot.Cmd) (string, error) {
	log.Debug("cmd.Args: %q", cmd.Args)

	if len(cmd.Args) == 1 && cmd.Args[0] == "stats" {
		kl, max_nicklen := rank()
		fstr := fmt.Sprintf("%s%d%s", "%-", max_nicklen, "s : %3d\n")
		str := fmt.Sprintf("Stats since %s:\n", botstart.Format(time.RFC3339))
		for _, kv := range kl {
			str += fmt.Sprintf(fstr, kv.Key, kv.Val)
		}
		return str, nil
	}

	// prevent ddos/spam
	if didTry[cmd.User.Nick] {
		return "", nil
	}

	t := time.Now()
	if t.Hour() == hour && t.Minute() == minute {
		if isFirst {
			scores[cmd.User.Nick] += 2
			isFirst = false
			time.AfterFunc(1*time.Minute, func() {
				isFirst = true
			})
		} else {
			scores[cmd.User.Nick]++
		}
		tempBlockUser(cmd.User.Nick)
		return fmt.Sprintf("Whoop! %s total score: %d\n", cmd.User.Nick, scores[cmd.User.Nick]), nil
	} else if t.Hour() == hour && t.Minute() == minute-1 {
		scores[cmd.User.Nick]--
		tempBlockUser(cmd.User.Nick)
		return fmt.Sprintf("Too early, sucker! %s: %d\n", cmd.User.Nick, scores[cmd.User.Nick]), nil
	} else if t.Hour() == hour && t.Minute() == minute+1 {
		scores[cmd.User.Nick]--
		tempBlockUser(cmd.User.Nick)
		return fmt.Sprintf("Too late, sucker! %s: %d\n", cmd.User.Nick, scores[cmd.User.Nick]), nil
	}

	return "", nil
}

func pickupEnv() {
	h := os.Getenv("LEETBOT_HOUR")
	m := os.Getenv("LEETBOT_MINUTE")

	var err error
	if h != "" {
		hour, err = strconv.Atoi(h)
		if err != nil {
			hour = DEF_HOUR
		}
	}
	if m != "" {
		minute, err = strconv.Atoi(m)
		if err != nil {
			minute = DEF_MINUTE
		}
	}
}

func init() {
	botstart = time.Now()
	scores = make(map[string]int)
	didTry = make(map[string]bool)

	pickupEnv()

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats]",
		leet)
}
