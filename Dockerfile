FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o wordle-agent ./cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/customer-profiler-agent .

EXPOSE 5001

CMD ["./customer-profiler-agent"]