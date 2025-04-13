FROM golang:1.24-alpine AS builder

WORKDIR /craftyproxy

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux

RUN go build -ldflags="-s -w" -o /craftyproxy/main cmd/reverse-proxy/main.go

FROM alpine:3.21

COPY --from=builder  /craftyproxy/main /craftyproxy/main

CMD [ "/craftyproxy/main" ]
