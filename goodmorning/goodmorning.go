package goodmorning

import (
	"math/rand"
	"regexp"

	"github.com/go-chat-bot/bot"
)

const (
	pattern = "(?i)\\bgo[o]*d\\b(morgen|morgon|morgan|morning)\\b"
)

var (
	re        = regexp.MustCompile(pattern)
	greetings = []string{
		"God morgen, kjære %s :) Håper dagen din blir strålende isfrisk!",
		"Hold kjeft, %s! Verden kan gå og brenne!",
		"Voldta meg i røven, %s, jeg er morgenkåt!",
		"Hva faen mener du med det, %s?? Ypper du bråk, eller?!?",
		"Ja, det skulle tatt seg ut, du %s...",
		"%s: Det sa mor di au...",
		"Oh-la-la! %s har tydeligvis fått se noe i natt ;)",
	}
)

func goodmorning(cmd *bot.PassiveCmd) (string, error) {
	if re.MatchString(cmd.Raw) {
		return greetings[rand.Intn(len(greetings))], nil
	}
	return "", nil
}

func init() {
	bot.RegisterPassiveCommand("goodmorning", goodmorning)
}
