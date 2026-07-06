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
COPY Makefile ./Makefile

RUN make protobuf 

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/gateway/main.go

FROM alpine:latest

WORKDIR /app

COPY cmd/gateway/config.yaml ./config.yaml
COPY --from=builder /bin/gateway /app/gateway

EXPOSE 80

ENTRYPOINT ["/app/gateway"]
