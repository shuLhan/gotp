## SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
## SPDX-License-Identifier: GPL-3.0-or-later

.PHONY: all test build install serve-doc

VERSION:=$(shell git describe --tags)
LDFLAGS:=-ldflags "-s -w -X 'git.sr.ht/~shulhan/gotp.Version=$(VERSION)'"

all: test build

test:
	CGO_ENABLED=1 go test -race -failfast -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.out

build:
	mkdir -p _sys/usr/bin/
	go build $(LDFLAGS) -o _sys/usr/bin/ ./cmd/...

install: build
	install -D _sys/usr/bin/gotp $(DESTDIR)/usr/bin/gotp
	install -Dm644 _sys/etc/bash_completion.d/gotp $(DESTDIR)/etc/bash_completion.d/gotp
	install -Dm644 COPYING $(DESTDIR)/usr/share/licenses/gotp/COPYING

serve-doc:
	ciigo serve _doc
