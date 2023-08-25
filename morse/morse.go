package morse

import (
	"fmt"
	"strings"

	m "github.com/alwindoss/morse"
	"github.com/go-chat-bot/bot"
)

const (
	plugin = "MorseConverter"
)

func ascii2morse(input string) (string, error) {
	h := m.NewHacker()
	morseCode, err := h.Encode(strings.NewReader(input))
	if nil != err {
		return "", err
	}
	return string(morseCode), nil
}

func morse2ascii(morse string) (string, error) {
	h := m.NewHacker()
	ascii, err := h.Decode(strings.NewReader(morse))
	if nil != err {
		return "", err
	}
	return string(ascii), nil
}

func tomorse(cmd *bot.Cmd) (string, error) {
	if len(cmd.Args) < 1 {
		return fmt.Sprintf("%s: No input. Usage: !tomorse <input>", plugin), nil
	}

	morse, err := ascii2morse(strings.Join(cmd.Args, " "))
	if nil != err {
		return "", err
	}
	return morse, nil
}

func frommorse(cmd *bot.Cmd) (string, error) {
	if len(cmd.Args) < 1 {
		return fmt.Sprintf("%s: No input. Usage: !frommorse <input>", plugin), nil
	}

	ascii, err := morse2ascii(strings.Join(cmd.Args, " "))
	if nil != err {
		return "", err
	}
	return ascii, nil
}

func init() {
	bot.RegisterCommand(
		"tomorse",
		"Convert ASCII input to morse",
		"<input>",
		tomorse,
	)
	bot.RegisterCommand(
		"frommorse",
		"Convert morse input to ASCII",
		"<input>",
		frommorse,
	)
}
