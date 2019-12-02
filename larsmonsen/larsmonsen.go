package larsmonsen

import (
	"os"
	"regexp"

	"github.com/go-chat-bot/bot"
	"github.com/oddlid/dvdgbot/quoteshuffle"
	log "github.com/sirupsen/logrus"
)

const (
	PLUGIN     = "LarsMonsen"
	FACTS_FILE = "/tmp/larsmonsenfacts.json"
	pattern    = "(?i)\\b(lars|monsen)\\b"
	cmdName    = "larsmonsen"
)

var (
	qd   *quoteshuffle.QuoteData
	re   = regexp.MustCompile(pattern)
	_log = log.WithField("plugin", PLUGIN)
)

func envDefStr(key, fallback string) string {
	val, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	return val // might still be empty, if set, but empty in ENV
}

func larsmonsen(command *bot.PassiveCmd) (string, error) {
	if re.MatchString(command.Raw) {
		return qd.Shuffle(), nil
	}
	return "", nil
}

func init() {
	var err error
	qd, err = quoteshuffle.New(envDefStr("LARSMONSENFACTS_FILE", FACTS_FILE))
	if err != nil {
		_log.WithError(err).Error("Error loading facts file")
		return
	}
	bot.RegisterPassiveCommand(cmdName, larsmonsen)
}
