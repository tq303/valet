DEST := $(HOME)/go/bin/val

.PHONY: install build test release

install:
	go build -o $(DEST) .
	chmod +x $(DEST)
	@echo "Installed to $(DEST)"

build:
	go build ./...

test:
	go test ./...

release:
	@bash scripts/release.sh
