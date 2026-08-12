.PHONY: build all clean

# Default to arm64 for Termux on modern Android devices
GOARCH ?= arm64
GOOS ?= android

build:
CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o terpodman .

all: build-arm64 build-arm

build-arm64:
CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build -ldflags="-s -w" -o terpodman-arm64 .

build-arm:
CGO_ENABLED=0 GOOS=android GOARCH=arm go build -ldflags="-s -w" -o terpodman-arm .

clean:
rm -f terpodman terpodman-arm64 terpodman-arm

install: build
cp terpodman $$PREFIX/bin/ 2>/dev/null || cp terpodman ~/bin/

test:
go test ./...
