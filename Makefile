GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: all
all:
	@$(MAKE) deps
	@$(MAKE) build

.PHONY: deps
deps:
	go mod download -x

.PHONY: build
build:
ifneq (windows, $(GOOS))
	go build -a -o build/bin/telegraf-output-kinesis-data-firehose ./cmd/
else
	go build -a -o build/bin/telegraf-output-kinesis-data-firehose.exe ./cmd/
endif
	cp plugin.conf build/bin/

.PHONY: test
test:
	go test -race -covermode=atomic -coverprofile=covprofile -coverprofile=coverage.out -parallel 4 -timeout 2h ./...

.PHONY: lint
lint:
	golangci-lint run -c ./.golangci.yml

.PHONY: tidy
tidy:
	go mod verify
	go mod tidy

.PHONY: clean
clean:
	rm -rf build
