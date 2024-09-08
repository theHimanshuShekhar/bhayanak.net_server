# syntax=docker/dockerfile:1

FROM golang:1.22.6-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o bchat-server cmd/server/main.go

EXPOSE 8080

CMD ["./bchat-server"]