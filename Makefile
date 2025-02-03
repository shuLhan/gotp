## SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
## SPDX-License-Identifier: GPL-3.0-or-later

.PHONY: all
all: test lint build

test:
	CGO_ENABLED=1 go test -race -failfast -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

.PHONY: lint
lint:
	go run ./internal/cmd/gocheck ./...
	go vet ./...

.PHONY: build
build:
	mkdir -p _sys/usr/bin/
	go build -o _sys/usr/bin/ ./cmd/...

.PHONY: install
install: build
	install -D _sys/usr/bin/gotp $(DESTDIR)/usr/bin/gotp
	install -Dm644 \
	        _sys/usr/share/bash-completion/completions/gotp \
	  $(DESTDIR)/usr/share/bash-completion/completions/gotp
	install -Dm644 COPYING $(DESTDIR)/usr/share/licenses/gotp/COPYING

.PHONY: install-darwin
install-darwin: DESTDIR=/usr/local
install-darwin: build
	install -D _sys/usr/bin/gotp $(DESTDIR)/bin/gotp
	install -Dm644 \
	        _sys/usr/share/bash-completion/completions/gotp \
	  $(DESTDIR)/etc/bash_completion.d/gotp
	install -Dm644 COPYING $(DESTDIR)/share/gotp/COPYING

.PHONY: serve-doc
serve-doc:
	ciigo serve _doc

.PHONY: uninstall-darwin
uninstall-darwin: DESTDIR=/usr/local
uninstall-darwin:
	rm -f $(DESTDIR)/etc/bash_completion.d/gotp
	rm -f $(DESTDIR)/share/gotp/COPYING
	rm -f $(DESTDIR)/bin/gotp
