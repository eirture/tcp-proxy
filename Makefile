
KODO_VERSION ?= unknown
BUILD_DATE ?= $(shell date +%Y-%m-%d)

GO_LDFLAGS := -X github.com/eirture/tcp-proxy/pkg/build.Version=$(KODO_VERSION) $(GO_LDFLAGS)
GO_LDFLAGS := -X github.com/eirture/tcp-proxy/pkg/build.BuildDate=$(BUILD_DATE) $(GO_LDFLAGS)


.PHONY: build
build:
	@mkdir -p bin
	go build -ldflags "${GO_LDFLAGS}" -v -o bin/ ./cmd/tcp-proxy

.PHONY: clean
clean:
	@rm -rf bin/*
