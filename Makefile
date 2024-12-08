CLI_IMAGE := mitch
CLICKHOUSE_IMAGE := clickhouse:lts-focal
CLICKHOUSE_CONTAINER := clickhouse-server
CLI_CONTAINER := mitch-runner
NETWORK_NAME := clickhouse-test
CLICKHOUSE_PORT := 9000
HOST_PORT := 9000

.PHONY: all setup build test clean stop

all: setup build test

setup:
	@echo "Setting up ClickHouse server..."
	@docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)
	@docker pull $(CLICKHOUSE_IMAGE)

build:
	@echo "Building `mitch` container..."
	@docker build -t $(CLI_IMAGE) .

test:
	@echo "Running commands and tests..."
	@docker run -d --name $(CLICKHOUSE_CONTAINER) \
		--network $(NETWORK_NAME) \
		-p $(HOST_PORT):$(CLICKHOUSE_PORT) \
		$(CLICKHOUSE_IMAGE)
	@docker run --rm --name $(CLI_CONTAINER) \
		--network $(NETWORK_NAME) \
		-e CLICKHOUSE_SERVER=$(CLICKHOUSE_CONTAINER) \
		$(CLI_IMAGE) /bin/sh -c "\
		go build -o /app/migrate ./cmd/migrate && \
		/app/migrate --host=$(CLICKHOUSE_CONTAINER) --port=$(CLICKHOUSE_PORT) test"

clean:
	@echo "Cleaning up containers and network..."
	@docker stop $(CLICKHOUSE_CONTAINER) || true
	@docker rm $(CLICKHOUSE_CONTAINER) || true
	@docker network rm $(NETWORK_NAME) || true

stop:
	@echo "Stopping running containers..."
	@docker stop $(CLICKHOUSE_CONTAINER) || true
