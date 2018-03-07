FROM alpine:latest
MAINTAINER Odd E. Ebbesen <oddebb@gmail.com>

RUN addgroup -g 1000 bajsbot && adduser -u 1000 -G bajsbot -D bajsbot

RUN apk add --no-cache --update \
		ca-certificates \
		tini \
		&& \
		rm -rf /var/cache/apk/*

COPY bajsbot.bin /usr/bin/bajsbot
RUN chown bajsbot:bajsbot /usr/bin/bajsbot && chmod 755 /usr/bin/bajsbot

USER bajsbot

ENTRYPOINT ["tini", "-g", "--", "bajsbot"]]
