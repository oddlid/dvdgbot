package leet

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-chat-bot/bot"
)

const (
	DEF_HOUR   int    = 13
	DEF_MINUTE int    = 37
	SCORE_FILE string = "/tmp/leetbot_scores.json"
)

var (
	isFirst    bool = true
	cacheDirty bool = false
	hour       int  = DEF_HOUR
	minute     int  = DEF_MINUTE
	callstack  int
	botstart   time.Time
	scores     map[string]int
	didTry     map[string]bool
	mx         sync.RWMutex
)

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

func load(r io.Reader) error {
	jb, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(jb, &scores)
}

func save(w io.Writer) (int, error) {
	jb, err := json.Marshal(scores)
	if err != nil {
		return 0, err
	}
	jb = append(jb, '\n')
	return w.Write(jb)
}

func savestats() {
	file, err := os.Create(SCORE_FILE)
	if err != nil {
		log.Errorf("Error opening %q for saving: %s", SCORE_FILE, err)
		return
	}
	defer file.Close()
	n, err := save(file)
	if err != nil {
		log.Errorf("Error saving json file %q: %s", SCORE_FILE, err)
		return
	}
	log.Infof("Saved %d bytes of scores to %q", n, SCORE_FILE)
}

func delayedSave() bool {
	mx.RLock()
	defer mx.RUnlock()
	// nothing has changed, so return false to say we did nothing
	if !cacheDirty {
		return false
	}

	// if we've gotten to this point, we should be within the 2 minute timeframe where
	// scores might be changed, so we wait 3 minutes just to be sure, and then save all in one go.
	// Otherwise, if there's a lot of users triggering at the same time, we do a lot more disk writes than we need.
	// Also, we try to limit the number of goroutines that will try to save, using a counter.
	callstack++
	if callstack == 1 {
		time.AfterFunc(3*time.Minute, func() {
			mx.Lock()
			if cacheDirty {
				savestats()
				cacheDirty = false
			}
			callstack = 0
			mx.Unlock()
		})
	}
	return true
}

// score adds/substracts a users points, then prevents them from entering more than once in 2 minutes
func score(nick string, points int) {
	if didTry[nick] {
		return
	}

	didTry[nick] = true
	mx.Lock()
	scores[nick] += points
	cacheDirty = true
	mx.Unlock()
	// reset the nick lock as soon as we're out of the accepted timeframe again
	time.AfterFunc(2*time.Minute, func() {
		mx.Lock()
		didTry[nick] = false
		mx.Unlock()
	})
}

func leet(cmd *bot.Cmd) (string, error) {
	log.Debugf("cmd.Args: %q", cmd.Args)

	if len(cmd.Args) == 1 && cmd.Args[0] == "stats" {
		kl, max_nicklen := rank()
		fstr := fmt.Sprintf("%s%d%s", "%-", max_nicklen, "s : %3d\n")
		str := fmt.Sprintf("Stats since %s:\n", botstart.Format(time.RFC3339))
		for _, kv := range kl {
			str += fmt.Sprintf(fstr, kv.Key, kv.Val)
		}
		return str, nil
	} else if len(cmd.Args) == 1 {
		return fmt.Sprintf("Unrecognized argument: %s. Usage: !1337 [stats]", cmd.Args[0]), nil
	}

	// prevent ddos/spam
	if didTry[cmd.User.Nick] {
		return "", nil
	}

	t := time.Now()
	if t.Hour() == hour && t.Minute() == minute {
		if isFirst {
			score(cmd.User.Nick, 2)
			mx.Lock()
			isFirst = false
			mx.Unlock()
			time.AfterFunc(2*time.Minute, func() {
				mx.Lock()
				isFirst = true
				mx.Unlock()
			})
		} else {
			score(cmd.User.Nick, 1)
		}
		return fmt.Sprintf("Whoop! %s total score: %d\n", cmd.User.Nick, scores[cmd.User.Nick]), nil
	} else if t.Hour() == hour && t.Minute() == minute-1 {
		score(cmd.User.Nick, -1)
		return fmt.Sprintf("Too early, sucker! %s: %d\n", cmd.User.Nick, scores[cmd.User.Nick]), nil
	} else if t.Hour() == hour && t.Minute() == minute+1 {
		score(cmd.User.Nick, -1)
		return fmt.Sprintf("Too late, sucker! %s: %d\n", cmd.User.Nick, scores[cmd.User.Nick]), nil
	}

	// score() sets cacheDirty, so if it's set, it means we matched within the timeframe above,
	// which means at least one score was modified, so we save.
	if cacheDirty {
		delayedSave()
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

	file, err := os.Open(SCORE_FILE)
	if err == nil {
		err := load(file)
		if err != nil {
			log.Errorf("Error loading config from %q: %s", SCORE_FILE, err)
		} else {
			log.Infof("Loaded scores from %q", SCORE_FILE)
		}
		file.Close()
	}

	bot.RegisterCommand(
		"1337",
		"Register 1337 event, or print stats",
		"[stats]",
		leet)
}
