FROM golang:alpine as builder

WORKDIR /src

COPY . .

RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o client

FROM alpine:latest


EXPOSE 8089

WORKDIR /app

COPY --from=builder /src/client .

ENTRYPOINT ["/app/client"]