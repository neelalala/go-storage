FROM golang:1.26 AS builder

RUN apt update && apt install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH="$PATH:$(go env GOPATH)/bin"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/gateway ./cmd/gateway
COPY internal/gateway ./internal/gateway
COPY pkg/proto/storage/storage.proto ./pkg/proto/storage/
COPY pkg/proto/metadata/metadata.proto ./pkg/proto/metadata/
COPY Makefile ./Makefile

RUN protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  pkg/proto/storage/storage.proto
RUN protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  pkg/proto/metadata/metadata.proto


RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/gateway/main.go

FROM alpine:latest

WORKDIR /app

COPY cmd/gateway/config.yaml ./config.yaml
COPY --from=builder /bin/gateway /app/gateway

EXPOSE 80

ENTRYPOINT ["/app/gateway"]
