BINARY=terraform-provider-asashv
VERSION?=0.1.0
OS_ARCH=linux_amd64
PLUGIN_DIR=$(HOME)/.terraform.d/plugins/registry.terraform.io/asashv/asashv/$(VERSION)/$(OS_ARCH)

.PHONY: fmt tidy build install test clean

fmt:
	gofmt -w main.go internal/client/*.go internal/provider/*.go

tidy:
	go mod tidy

build: fmt tidy
	go build -o $(BINARY)

install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(BINARY) $(PLUGIN_DIR)/$(BINARY)_v$(VERSION)
	chmod +x $(PLUGIN_DIR)/$(BINARY)_v$(VERSION)

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/
