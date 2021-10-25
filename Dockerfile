FROM golang:1.17-alpine AS build

WORKDIR /go/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /go/bin/dmn.bin ./cmd/dota_market_notifier

FROM alpine:latest

COPY --from=build /go/bin/dmn.bin /dmn.bin

ENTRYPOINT ["/dmn.bin"]
