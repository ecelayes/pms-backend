# --- 1. Base Stage (Común) ---
FROM golang:1.25-alpine AS base
WORKDIR /app

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

# --- 2. Development Stage (Target: dev) ---
FROM base AS dev

RUN go install github.com/air-verse/air@latest

COPY . .

CMD ["air", "-c", ".air.toml"]

# --- 3. Builder Stage ---
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o pms-backend cmd/api/main.go

# --- 4. Production Stage (Target: prod) ---
FROM alpine:3.19 AS prod

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/pms-backend .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/scripts ./scripts

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080

CMD ["./pms-backend"]
