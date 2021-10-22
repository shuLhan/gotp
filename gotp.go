// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotp

import (
	"strings"
	"unicode"
)

// List of available algorithm for Provider.
const (
	HashSHA1   = "SHA1" // Default algorithm.
	HashSHA256 = "SHA256"
	HashSHA512 = "SHA512"
)

const (
	configDir   = "gotp"
	configFile  = "gotp.conf"
	defaultHash = HashSHA1

	// List of known providers
	providerNameAegis = "aegis"
)

//
// normalizeLabel convert non alpha number, hyphen, underscore, or period
// characters into "-".
//
func normalizeLabel(in string) (out string) {
	var (
		buf         strings.Builder
		replacement rune = '-'
	)
	for _, r := range in {
		if unicode.IsLetter(r) || unicode.IsDigit(r) ||
			r == '-' || r == '_' || r == '.' {
			buf.WriteRune(r)
		} else {
			buf.WriteRune(replacement)
		}
	}
	return strings.ToLower(buf.String())
}
