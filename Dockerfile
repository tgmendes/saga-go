# syntax=docker/dockerfile:1
FROM golang:1.17-alpine
ARG service
WORKDIR /app

COPY . .
COPY go.mod ./
COPY go.sum ./

RUN go mod download

WORKDIR /app/${service}

RUN go build -o /saga-app

CMD ["/saga-app"]
