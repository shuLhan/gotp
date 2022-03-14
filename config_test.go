// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewConfig(t *testing.T) {
	cases := []struct {
		desc       string
		configFile string
		expConfig  *config
		expError   string
	}{{
		desc:       "With openssh rsa",
		configFile: "testdata/rsa.conf",
		expConfig: &config{
			PrivateKey: "testdata/rsa",
			Issuers: map[string]string{
				"email-domain": "XYZ",
				"test":         "ABCD",
			},
			file: "testdata/rsa.conf",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		gotConfig, err := newConfig(c.configFile)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		test.Assert(t, "Issuer", c.expConfig, gotConfig)
	}
}
