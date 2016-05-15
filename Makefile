#BIRDCATCHER_VERSION?=$(shell git describe --tags)

all: bin/birdcatcher

bin:
	mkdir -p bin

bin/birdcatcher: bin */*.go
	go build -o bin/birdcatcher cmd/birdcatcher.go

get:
	go get -t ./...

fmt:
	go fmt ./...

install: all
	cp bin/* $(DESTDIR)/usr/bin

develop: all
	ln -f -s `pwd`/bin/* -t /usr/local/bin/

test:
	@test -z "(shell find . -name '*.go' | xargs gofmt -l)" || (echo "Need to run 'go fmt ./...'"; exit 1)
	go vet ./...
	go test ./...
