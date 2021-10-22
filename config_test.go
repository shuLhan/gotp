// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
