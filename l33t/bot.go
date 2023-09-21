package l33t

import (
	"errors"

	"github.com/go-chat-bot/bot"
)

var (
	errNilChatBot   = errors.New("chatBot is nil")
	errEmptyChannel = errors.New("channel name is empty")
	errEmptyMsg     = errors.New("message is empty")
	errNilReceiver  = errors.New("receiver is nil")
)

type Bot struct {
	chatBot *bot.Bot
}

func (b *Bot) sendMessage(channel, message string) error {
	if b == nil {
		return errNilReceiver
	}
	if b.chatBot == nil {
		return errNilChatBot
	}
	if channel == "" {
		return errEmptyChannel
	}
	if message == "" {
		return errEmptyMsg
	}

	b.chatBot.SendMessage(
		bot.OutgoingMessage{
			Target:  channel,
			Message: message,
			Sender:  &bot.User{},
		},
	)

	return nil
}
