BINARY_NAME := cherrypicker
GO := go
GOFLAGS :=
LDFLAGS :=

.PHONY: build clean tidy fmt vet

build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .

clean:
	rm -rf bin/

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...
