FROM golang:1.22.2-buster as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mitch .
RUN go test -v ./...

FROM debian:buster-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/mitch /usr/local/bin/mitch
COPY --from=builder /app/testdata/mitch.env /usr/local/bin/testdata/mitch.env
ENTRYPOINT ["/usr/local/bin/mitch"]
