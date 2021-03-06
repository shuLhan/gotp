## SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
## SPDX-License-Identifier: GPL-3.0-or-later
.PHONY: all install

all:
	CGO_ENABLED=1 go test -race -failfast ./...

install:
	go install ./cmd/gotp
