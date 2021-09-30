

.PHONY: build
build:
	@mkdir -p bin
	go build -v -o bin/ ./cmd/tcp-proxy

.PHONY: clean
clean:
	@rm -rf bin/*
