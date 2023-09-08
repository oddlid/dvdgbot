package larsmonsen

import (
	"regexp"

	"github.com/go-chat-bot/bot"
	"github.com/rs/zerolog/log"

	"github.com/oddlid/dvdgbot/quoteshuffle"
	"github.com/oddlid/dvdgbot/util"
)

type LarsMonsen struct {
	qd *quoteshuffle.QuoteData
	rx *regexp.Regexp
}

const (
	DefaultPattern     = `(?i)\\b(lars|monsen)\\b`
	DefaultCommandName = `larsmonsen`
	factsFile          = `/tmp/larsmonsenfacts.json`
)

func New(fileName, pattern string) (*LarsMonsen, error) {
	qd, err := quoteshuffle.New(fileName)
	if err != nil {
		return nil, err
	}

	rx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &LarsMonsen{
		qd: qd,
		rx: rx,
	}, nil
}

func (lm *LarsMonsen) Quote(command *bot.PassiveCmd) (string, error) {
	if lm.rx.MatchString(command.Raw) {
		return lm.qd.Shuffle()
	}
	return "", nil
}

func init() {
	if lm, err := New(util.EnvDefStr("LARSMONSENFACTS_FILE", factsFile), DefaultPattern); err != nil {
		log.Error().Err(err).Send()
	} else {
		bot.RegisterPassiveCommand(DefaultCommandName, lm.Quote)
	}
}
