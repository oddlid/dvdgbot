package timestamp

/*
Super simple bot function that just registers a detailed timestamp for when a
message was received, then prints timestamp and message.
Odd, 2019-03-17 17:07:59
*/

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-chat-bot/bot"
)

const (
	DefaultCommandName = `ts`
	Description        = `Prepend a message with a detailed timestamp`
	Params             = `[message]`
)

func Prepend(cmd *bot.Cmd) (string, error) {
	t := time.Now()
	ts := fmt.Sprintf("[%02d:%02d:%02d:%09d]", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	return fmt.Sprintf("%s <%s>: %s", ts, cmd.User.Nick, strings.Join(cmd.Args[0:len(cmd.Args)], " ")), nil
}
