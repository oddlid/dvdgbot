package l33t

import (
	"testing"

	"github.com/go-chat-bot/bot"
	"github.com/stretchr/testify/assert"
)

func Test_Bot_sendMessage_whenReceiverIsNil(t *testing.T) {
	t.Parallel()
	assert.ErrorIs(t, (*Bot)(nil).sendMessage("", ""), errNilReceiver)
}

func Test_Bot_sendMessage_whenChatBotIsNil(t *testing.T) {
	t.Parallel()
	b := Bot{}
	assert.ErrorIs(t, b.sendMessage("", ""), errNilChatBot)
}

func Test_Bot_sendMessage_whenEmptyChannel(t *testing.T) {
	t.Parallel()
	b := Bot{
		chatBot: &bot.Bot{},
	}
	assert.ErrorIs(t, b.sendMessage("", ""), errEmptyChannel)
}

func Test_Bot_sendMessage_whenEmptyMessage(t *testing.T) {
	t.Parallel()
	b := Bot{
		chatBot: &bot.Bot{},
	}
	assert.ErrorIs(t, b.sendMessage("channel", ""), errEmptyMsg)
}

func Test_Bot_sendMessage(t *testing.T) {
	t.Parallel()
	b := Bot{
		chatBot: &bot.Bot{},
	}
	assert.Panics(t, func() {
		b.sendMessage("chan", "msg")
	})
}
