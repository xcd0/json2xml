BIN           := ./json2xml.exe
VERSION       := `git tag -l | sort -rV | head -n1`
REVISION      := `git rev-parse --short HEAD`
FLAG          := -ldflags='-X main.version='$(VERSION)' -X main.revision='$(REVISION)' -s -w -extldflags="-static" -buildid=' -a -tags netgo -installsuffix -trimpath

all:
	cat ./makefile

build:
	rm -rf ./files
	make generate
	make fmt
	go build

release:
	rm -rf ./files
	make generate
	make fmt
	go build $(FLAG)
	make upx 
	@echo Success!

fmt:
	goimports -w *.go
	gofmt -w *.go

generate:
	go generate

upx:
	upx --lzma $(BIN)

