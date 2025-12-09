FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pms-backend cmd/api/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/pms-backend .

COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/scripts ./scripts

EXPOSE 8080

CMD ["./pms-backend"]
