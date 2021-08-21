ARG BUILD_BIN=dvdgbot.bin
ARG BOT_USER=dvdgbot

FROM golang:latest AS builder
# Specifying the same ARG without value copies it into this FROM-scope, from the outside scoped ARG before FROM
ARG BUILD_BIN
# Set timezone so that the compiled date will be correct
ARG ARG_TZ=Europe/Stockholm
ENV TZ=${ARG_TZ}
# Note that $GOPATH is set to /go from the parent image
COPY . ${GOPATH}/src/github.com/oddlid/dvdgbot/
WORKDIR ${GOPATH}/src/github.com/oddlid/dvdgbot/
# Since this repo is now in Go modules format, we should be able to just build, and skip go get first
RUN make BINARY=${BUILD_BIN}


# We don't do anything about TZ in the image for running the app, as one would rather specify
# that at runtime, and not have it baked into the image

FROM alpine:latest
LABEL maintainer="Odd E. Ebbesen <oddebb@gmail.com>"
# Specifying the same ARG without value copies it into this FROM-scope, from the outside scoped ARG before FROM
ARG BUILD_BIN
ARG BOT_USER

RUN addgroup -g 1000 ${BOT_USER} && adduser -u 1000 -G ${BOT_USER} -D ${BOT_USER}

RUN apk add --no-cache --update ca-certificates \
		&& rm -rf /var/cache/apk/*

COPY --from=builder /go/src/github.com/oddlid/dvdgbot/${BUILD_BIN} /usr/local/bin/
RUN chown ${BOT_USER}:${BOT_USER} /usr/local/bin/${BUILD_BIN} && chmod 555 /usr/local/bin/${BUILD_BIN}

USER ${BOT_USER}
ENV BOT_BIN=${BUILD_BIN}
# Using the shell format of CMD is most useful here, as variable expansion then works as intended
CMD ${BOT_BIN} -h
