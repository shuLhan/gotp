.PHONY: all install

all:
	CGO_ENABLED=1 go test -race -failfast ./...

install:
	go install ./cmd/gotp
