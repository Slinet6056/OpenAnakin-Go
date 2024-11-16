FROM golang:1.23.3-alpine AS builder

LABEL org.opencontainers.image.source="https://github.com/Slinet6056/OpenAnakin-Go"
LABEL org.opencontainers.image.licenses="MIT"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o openanakin-go cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/openanakin-go .

COPY config.yaml .

EXPOSE 8080

CMD ["./openanakin-go"]
