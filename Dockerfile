FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
# COPY go.sum ./ # No dependencies yet

COPY . .

RUN go mod tidy
RUN go build -o rae-server cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/rae-server .

EXPOSE 8080

CMD ["./rae-server"]
