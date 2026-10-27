# ==========================================
# Step 1: Build Stage (Go Compiler)
# ==========================================
FROM golang:alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o krypto ./Backend

# ==========================================
# Step 2: Final Minimal Runtime Stage
# ==========================================
FROM alpine:3.20

RUN apk add --no-cache \
    ffmpeg \
    ca-certificates \
    tzdata

WORKDIR /app

COPY --from=builder /app/krypto /app/krypto
COPY --from=builder /app/Frontend /app/Frontend

EXPOSE 2409

ENV GIN_MODE=release
ENV PORT=2409

ENTRYPOINT ["/app/krypto"]