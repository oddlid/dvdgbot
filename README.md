# dvdgbot

This is my personal playground for IRC bot stuff in Go.
It will probably not be directly usable for others, but one might find bits and pieces that can be modified and reused.

It currently has some submodules that can be enabled/disabled by imports in main.go:

- *leet*:
  * This is the main motivation for the whole bot. It's a game.
  * Triggered by: `!1337 [stats|reload]`
  * See separate documentation.
- *quoteshuffle*:
  * Takes a JSON file with quotes (or whatever strings), and returns a random quote from the list.
  * To avoid often getting the same quotes, it will move the returned quote from the `src` array to the `dst`array, and when the `src` array is empty, all quotes are moved back to the `src` array from the `dst`array. This way, you won't see the same quote again until all others have been shown.
  * JSON format:
      `
      {
		"src": [
			"quote one",
			"quote two",
			"..."
		],
		"dst": [
		]
	  }
      `
- *larsmonsen*:
  * Based on https://github.com/go-chat-bot/plugins/chucknorris but has text from larsmonsenfacts.com
  * Text is saved in a separate JSON file, shuffled and rotated by the `quoteshuffle`module.
  * Triggered by the words "lars" or "monsen" (case insensitive) anywhere in a channel message.
- *timestamp*:
  * Just prepends a detailed timestamp to a message.
  * Triggered by the prefix command `!ts`
- *tool*:
  * This is just a helper for Makefile, to make it easier to get the same date format both on Linux/OSX/Windows. Not used by the bot itself.
- *userwatch*:
  * Lets you add welcome/bye messages for given nicks for JOIN/PART/QUIT.
  * See separate documentation.
- *xkcdbot*:
  * Returns the image URL for an XKCD comic.
  * Triggered by: `!xkcd get <ID>|latest`
