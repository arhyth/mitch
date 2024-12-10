CLI_IMAGE := mitch:test
CLICKHOUSE_IMAGE := clickhouse:lts-focal
CLICKHOUSE_CONTAINER := clickhouse-server
CLI_CONTAINER := mitch-test-runner
NETWORK_NAME := clickhouse-test
CLICKHOUSE_PORT := 9000
HOST_PORT := 9000

.PHONY: setup build test clean

rerun: clean build test

setup:
	@echo "Setting up network and images..."
	@docker network inspect $(NETWORK_NAME) >/dev/null 2>&1 || docker network create $(NETWORK_NAME)
	@docker pull $(CLICKHOUSE_IMAGE)

build:
	@echo "Building $(CLI_IMAGE) container..."
	@docker build --rm -f ./Dockerfile-test -t $(CLI_IMAGE) .

test:
	@echo "Running containers..."
	@docker run -d --rm --name $(CLICKHOUSE_CONTAINER) \
		--network $(NETWORK_NAME) \
		-p $(HOST_PORT):$(CLICKHOUSE_PORT) \
		$(CLICKHOUSE_IMAGE)
	@docker run --rm --name $(CLI_CONTAINER) \
		--network $(NETWORK_NAME) \
		-e CLICKHOUSE_HOST=$(CLICKHOUSE_CONTAINER) \
		$(CLI_IMAGE)

clean:
	@echo "Cleaning up containers..."
	@if [ -n "$$(docker ps -aq --filter "name=$(CLICKHOUSE_CONTAINER)")" ]; then \
		docker stop $(CLICKHOUSE_CONTAINER); \
	fi
	@if [ -n "$$(docker images -q $(CLI_IMAGE))" ]; then \
		docker rmi $(CLI_IMAGE); \
	fi
