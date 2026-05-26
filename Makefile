DEST := $(HOME)/.local/bin/val

.PHONY: install build test

install:
	go build -o $(DEST) .
	chmod +x $(DEST)
	@echo "Installed to $(DEST)"

build:
	go build ./...

test:
	go test ./...
