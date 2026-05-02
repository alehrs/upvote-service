FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o upvote-service ./cmd/upvote-service/

FROM alpine:3.21
RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=builder /app/upvote-service .
EXPOSE 8080
ENTRYPOINT ["./upvote-service"]
