FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/migrate ./cmd/migrate

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrate ./cmd/migrate/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /bin/migrate /app/migrate
COPY migrations/metadata ./migrations/metadata

ENTRYPOINT ["/app/migrate"]
CMD ["up"]
