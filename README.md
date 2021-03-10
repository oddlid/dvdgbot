# dvdgbot

This is my personal playground for IRC bot stuff in Go.
It will probably not be directly usable for others, but one might find bits and pieces that can be modified and reused.

It currently has some submodules that can be enabled/disabled by imports in main.go:

- **leet**:
  * This is the main motivation for the whole bot. It's a game.
  * Triggered by: `!1337 [stats|reload]`
  * See separate documentation.
- **quoteshuffle**:
  * Takes a JSON file with quotes (or whatever strings), and returns a random quote from the list.
  * To avoid often getting the same quotes, it will move the returned quote from the `src` array to the `dst`array, and when the `src` array is empty, all quotes are moved back to the `src` array from the `dst`array. This way, you won't see the same quote again until all others have been shown.
  * JSON format:

      `{
		"src": [
			"quote one",
			"quote two",
			"..."
		],
		"dst": [
		]
	  }`

- **larsmonsen**:
  * Based on https://github.com/go-chat-bot/plugins/chucknorris but has text from larsmonsenfacts.com
  * Triggered by the words "lars" or "monsen" (case insensitive) anywhere in a channel message.
  * Text is saved in a separate JSON file, shuffled and rotated by the `quoteshuffle`module.
  * Defaults to loading facts from `/tmp/larsmonsenfacts.json`, but can be overridden by the env var `LARSMONSENFACTS_FILE`
- **timestamp**:
  * Just prepends a detailed timestamp to a message.
  * Triggered by the prefix command `!ts`
- *tool*:
  * This is just a helper for Makefile, to make it easier to get the same date format both on Linux/OSX/Windows. Not used by the bot itself.
- **userwatch**:
  * Lets you add welcome/bye messages for given nicks for JOIN/PART/QUIT.
  * Loads and saves config in `/tmp/userwatch.json`. Override not implemented yet.
  * See separate documentation.
- **xkcdbot**:
  * Returns the image URL for an XKCD comic.
  * Triggered by: `!xkcd get <ID>|latest`

## Building

First, get your local copy of this repo:

`$ git clone https://github.com/oddlid/dvdgbot.git`

`$ cd dvdgbot`

Then, you can build a binary directly on your machine, if you have Golang installed, or you can build in Docker.

### Build locally

`$ make`

Or, if you want the binary to be named something else than `dvdgbot.bin`, run:

`$ make BINARY=mycoolbot` (replace "mycoolbot" with desired name)

### Build in Docker

If you've already built a binary with `make` as above, and would like to make a Docker image out of it, do like this, assuming you go with the default binary name:

`$ docker build -t <repo>/<image_name>:<tag> -f Dockerfile.local .`

Or, if you built with another binary name:

`$ docker build -t <repo>/<image_name>:<tag> --build-arg BUILD_BIN=mycoolbot -f Dockerfile.local`

If you just want to build a Docker image directly, without the intermediate step of a local binary, run:

`$ docker build -t <repo>/<image_name>:<tag> .`

Or if you want the binary to have some other name inside the Docker image, run:

`$ docker build -t <repo>/<image_name>:<tag> --build-arg BUILD_BIN=mycoolbot .`

