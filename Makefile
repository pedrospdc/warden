VERSION := $(shell cat VERSION)

.PHONY: build build-windows test lint clean

build:
	go build ./...

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		go build -ldflags="-H windowsgui" -o warden.exe .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f warden.exe
