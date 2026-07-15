FROM golang:1.26 AS builder

RUN apt update && apt install -y protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH="$PATH:$(go env GOPATH)/bin"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/metadata ./cmd/metadata
COPY internal/metadata ./internal/metadata
COPY pkg/proto/storage/storage.proto ./pkg/proto/storage/
COPY pkg/proto/metadata/metadata.proto ./pkg/proto/metadata/
COPY migrations/metadata ./migrations/metadata

RUN protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  pkg/proto/storage/storage.proto
RUN protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  pkg/proto/metadata/metadata.proto 

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/metadata ./cmd/metadata/main.go

FROM alpine:latest

WORKDIR /app

COPY cmd/metadata/config.yaml ./config.yaml
COPY --from=builder /bin/metadata /app/metadata
COPY --from=builder /app/migrations/metadata ./migrations/metadata

EXPOSE 50051

ENTRYPOINT ["/app/metadata"]
