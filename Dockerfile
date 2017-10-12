FROM alpine:3.6
EXPOSE 8080

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/trello /usr/local/bin/trello

CMD /usr/local/bin/trello
