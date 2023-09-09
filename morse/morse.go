package morse

import (
	"fmt"
	"strings"

	m "github.com/alwindoss/morse"
	"github.com/go-chat-bot/bot"
)

const (
	plugin    = `MorseConverter`
	A2MCmd    = `tomorse`
	A2MDesc   = `Convert ASCII input to morse`
	A2MParams = `<input>`
	M2ACmd    = `frommorse`
	M2ADesc   = `Convert morse input to ASCII`
	M2AParams = `<input>`
)

type Bot struct {
	h m.Hacker
}

func NewBot() *Bot {
	return &Bot{
		h: m.NewHacker(),
	}
}

func (b *Bot) ascii2morse(input string) (string, error) {
	morseCode, err := b.h.Encode(strings.NewReader(input))
	if err != nil {
		return "", err
	}
	return string(morseCode), nil
}

func (b *Bot) morse2ascii(morse string) (string, error) {
	ascii, err := b.h.Decode(strings.NewReader(morse))
	if err != nil {
		return "", err
	}
	return string(ascii), nil
}

func usage(cmd, params string) string {
	return fmt.Sprintf("%s: No input. Usage: !%s %s", plugin, cmd, params)
}

func (b *Bot) ToMorse(cmd *bot.Cmd) (string, error) {
	if len(cmd.Args) < 1 {
		return usage(A2MCmd, A2MParams), nil
	}
	morse, err := b.ascii2morse(strings.Join(cmd.Args, " "))
	if err != nil {
		return "", err
	}
	return morse, nil
}

func (b *Bot) FromMorse(cmd *bot.Cmd) (string, error) {
	if len(cmd.Args) < 1 {
		return usage(M2ACmd, M2AParams), nil
	}
	ascii, err := b.morse2ascii(strings.Join(cmd.Args, " "))
	if err != nil {
		return "", err
	}
	return ascii, nil
}
