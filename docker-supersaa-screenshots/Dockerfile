FROM golang:latest as build

WORKDIR /go/src/app

COPY ./config.toml .
COPY ./main.go .
RUN mkdir config
COPY ./config/config.go config

RUN go mod init supersaa; go mod tidy
RUN go build -ldflags="-s -w" -trimpath

FROM chromedp/headless-shell:latest
RUN apt-get update; apt install dumb-init -y
RUN rm -rf /var/lib/apt/lists/*

RUN mkdir /tmp/screenshot
ENTRYPOINT ["dumb-init", "--"]

COPY --from=build /go/src/app/supersaa /tmp
COPY --from=build /go/src/app/config.toml /tmp

CMD ["/tmp/supersaa"]
