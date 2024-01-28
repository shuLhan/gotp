// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

// Package gotp core library for building gotp CLI.
package gotp

import (
	"strings"
	"unicode"
)

// List of available algorithm for Provider.
const (
	HashSHA1    = `SHA1` // Default algorithm.
	HashSHA256  = `SHA256`
	HashSHA512  = `SHA512`
	defaultHash = HashSHA1
)

const (
	configFile     = `gotp.conf`
	privateKeyFile = `gotp.key`
)

// List of known providers.
const (
	providerNameAegis = `aegis`
)

// Version define the latest version of this module and gotp CLI.
var Version = `0.4.0`

// normalizeLabel convert non alpha number, hyphen, underscore, or period
// characters into `-`.
func normalizeLabel(in string) (out string) {
	var (
		replacement = '-'

		buf strings.Builder
		r   rune
	)
	for _, r = range in {
		if unicode.IsLetter(r) || unicode.IsDigit(r) ||
			r == '-' || r == '_' || r == '.' {
			buf.WriteRune(r)
		} else {
			buf.WriteRune(replacement)
		}
	}
	return strings.ToLower(buf.String())
}
