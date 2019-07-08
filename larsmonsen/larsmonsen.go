package larsmonsen

import (
	"regexp"

	"github.com/go-chat-bot/bot"
	"github.com/oddlid/dvdgbot/quoteshuffle"
)

const (
	FACTS_FILE string = "/tmp/larsmonsenfacts.json"
	pattern           = "(?i)\\b(lars|monsen)\\b"
)

var (
	re = regexp.MustCompile(pattern)
	qd *quoteshuffle.QuoteData
)

func larsmonsen(command *bot.PassiveCmd) (string, error) {
	if re.MatchString(command.Raw) {
		return qd.Shuffle(), nil
	}
	return "", nil
}

func init() {
	qd = quoteshuffle.New(FACTS_FILE)
	bot.RegisterPassiveCommand("larsmonsen", larsmonsen)
}
