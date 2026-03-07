VERSION := $(shell cat VERSION)

WIN_DOCS := /mnt/c/Users/pedro/Documents

.PHONY: build build-windows deploy test lint clean

build:
	go build ./...

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		go build -ldflags="-H windowsgui" -o warden.exe .

deploy: build-windows
	cp warden.exe "$(WIN_DOCS)/warden.exe"

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f warden.exe
