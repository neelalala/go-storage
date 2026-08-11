FROM golang:1.26 AS builder

RUN apt update && apt install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH="$PATH:$(go env GOPATH)/bin"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/users ./cmd/users
COPY internal/users ./internal/users
COPY pkg/proto/users/users.proto ./pkg/proto/users/

RUN protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  pkg/proto/users/users.proto

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/users ./cmd/users/main.go

FROM alpine:latest

WORKDIR /app

COPY cmd/users/config.yaml ./config.yaml
COPY --from=builder /bin/users /app/users

EXPOSE 50051

ENTRYPOINT ["/app/users"]
