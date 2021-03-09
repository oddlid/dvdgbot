package goodmorning

import (
	"fmt"
	"math/rand"
	"regexp"

	"github.com/go-chat-bot/bot"
)

const (
	pattern = "(?i)\\bgo[o]*d\\s+(morg[eoa]n|morron|morning)\\b"
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
		"Ka i hæilvettes steike sattans svarte gampræva du meine med det, %s?!",
		"Mums, mums, slikk meg på nesa, %s ;)",
		"Sug rull, %s!",
	}
)

func goodmorning(cmd *bot.PassiveCmd) (string, error) {
	if re.MatchString(cmd.Raw) {
		return fmt.Sprintf(greetings[rand.Intn(len(greetings))], cmd.User.Nick), nil
	}
	return "", nil
}

func init() {
	bot.RegisterPassiveCommand("goodmorning", goodmorning)
}
