// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strconv"
)

func parseProviderAegis(file string) (issuers []*Issuer, err error) {
	logp := "parseProviderAegis"

	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	lines := bytes.Split(b, []byte("\n"))
	for x, line := range lines {
		u, err := url.Parse(string(line))
		if err != nil {
			return nil, fmt.Errorf("%s: line %d: invalid format %q", logp, x, line)
		}
		if u.Host != "totp" {
			continue
		}

		q := u.Query()
		issuer := &Issuer{
			Label:  normalizeLabel(u.Path[1:]),
			Hash:   q.Get("algorithm"),
			Secret: q.Get("secret"),
			Name:   q.Get("issuer"),
		}

		val := q.Get("digits")
		issuer.Digits, err = strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("%s: line %d: invalid digits %q",
				logp, x, val)
		}

		val = q.Get("period")
		issuer.TimeStep, err = strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("%s: line %d: invalid period %q",
				logp, x, val)
		}

		issuers = append(issuers, issuer)
	}
	return issuers, nil
}
