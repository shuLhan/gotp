.PHONY: all install

all:
	go test -race -failfast ./...

install:
	go install ./cmd/gotp
