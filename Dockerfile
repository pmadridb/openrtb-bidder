# Stage 1: Build binary
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy all project files
COPY . .

# Build the entire package module (./cmd/bidder)
RUN CGO_ENABLED=0 GOOS=linux go build -o bidder ./cmd/bidder

# Stage 2: Minimal runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -g '' appuser
USER appuser

WORKDIR /home/appuser/

COPY --from=builder /app/bidder .

EXPOSE 8080

ENTRYPOINT ["./bidder"]