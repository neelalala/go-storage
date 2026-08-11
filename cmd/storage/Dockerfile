FROM golang:1.26 AS builder

RUN apt update && apt install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH="$PATH:$(go env GOPATH)/bin"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/storage ./cmd/storage
COPY internal/storage ./internal/storage
COPY pkg/proto/storage/storage.proto ./pkg/proto/storage/
COPY Makefile ./Makefile

RUN protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  pkg/proto/storage/storage.proto

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/storage ./cmd/storage/main.go

FROM alpine:latest

WORKDIR /app

COPY cmd/storage/config.yaml ./config.yaml
COPY --from=builder /bin/storage /app/storage

EXPOSE 50051

ENTRYPOINT ["/app/storage"]
