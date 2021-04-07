# syntax = docker/dockerfile:1.0.0-experimental
FROM golang:1.15.2 as builder

ENV GO111MODULE on
ENV GOPRIVATE "bitbucket.org/latonaio"

WORKDIR /go/src/bitbucket.org/latonaio

COPY go.mod .

RUN git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"

RUN mkdir /root/.ssh/ && touch /root/.ssh/known_hosts && ssh-keyscan -t rsa bitbucket.org >> /root/.ssh/known_hosts

RUN --mount=type=secret,id=ssh,target=/root/.ssh/id_rsa go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o aion-send-test-kube .

# Runtime Container
FROM alpine:3.12

RUN apk update \
 && apk add --no-cache \
            alsa-utils \
            pulseaudio socat

ENV SERVICE=:aion-send-test-kube \
    POSITION=Runtime \
    AION_HOME="/var/lib/aion" \
    APP_DIR="${AION_HOME}/${POSITION}/${SERVICE}"

WORKDIR ${APP_DIR}


COPY --from=builder /go/src/bitbucket.org/latonaio/aion-send-test-kube .

CMD ["./aion-send-test-kube"]
