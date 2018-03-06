package main

import (
	"github.com/go-chat-bot/bot/irc"
	_ "github.com/go-chat-bot/plugins/chucknorris"
	_ "github.com/oddlid/dvdgbot/larsmonsen"
)

func main() {
	irc.Run(&irc.Config{
		Server:   "irc.oftc.net",
		Channels: "#dvdg",
		User:     "dvdgbot",
		Nick:     "dvdgbot",
		Password: "",
		UseTLS:   true,
		Debug:    true,
	})
}
