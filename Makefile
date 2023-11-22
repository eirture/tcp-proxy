
VERSION ?= unknown
BUILD_DATE ?= $(shell date +%Y-%m-%d)

GO_LDFLAGS := -X github.com/eirture/tcp-proxy/pkg/build.Version=$(VERSION) $(GO_LDFLAGS)
GO_LDFLAGS := -X github.com/eirture/tcp-proxy/pkg/build.Date=$(BUILD_DATE) $(GO_LDFLAGS)


.PHONY: build
build:
	@mkdir -p bin
	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "${GO_LDFLAGS}" -v -o bin/ ./cmd/tcp-proxy

.PHONY: pkg
pkg: clean
	make build GOOS=darwin && cd bin/ && zip tcp-proxy-${VERSION}-darwin-amd64.zip tcp-proxy
	cd ../
	make build GOOS=linux && cd bin/ && tar zcf tcp-proxy-${VERSION}-linux-amd64.tgz tcp-proxy

.PHONY: clean
clean:
	@rm -rf bin/*

.PHONY: install
install:
	go install -ldflags "${GO_LDFLAGS}" -v ./cmd/tcp-proxy

.PHONY: uninstall
uninstall:
	rm -f $(GOPATH)/bin/tcp-proxy

linux: GOOS=linux
linux: GOARCH=amd64
linux: build

.PHONY: docker
docker: linux
	docker build -t eirture/tcp-proxy:$(shell date +%Y%m%d) .

