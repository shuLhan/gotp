// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gotp

import (
	"fmt"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestCli_inputPrivateKey(t *testing.T) {
	cli := &Cli{
		cfg: &config{
			file:       "testdata/save.conf",
			isNotExist: true,
		},
	}

	cases := []struct {
		desc       string
		privateKey string
		exp        string
	}{{
		desc: "Without private key",
		exp:  "[gotp]\nprivate_key =\n",
	}, {
		desc:       "With private key",
		privateKey: "testdata/rsa",
	}}

	for _, c := range cases {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprintf(w, "%s\n", c.privateKey)

		gotPrivateKeyFile, err := cli.inputPrivateKey(r)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, cli.cfg.file, c.privateKey, gotPrivateKeyFile)
	}
}

func TestCli_Add(t *testing.T) {
	cli := &Cli{
		cfg: &config{
			Issuers: make(map[string]string),
			file:    "testdata/add.conf",
		},
	}

	err := cli.cfg.save()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc      string
		issuer    *Issuer
		expError  string
		expConfig string
	}{{
		desc:      "With nil issuer",
		expConfig: "[gotp]\nprivate_key =\n",
	}, {
		desc: "With invalid label",
		issuer: &Issuer{
			Label: "Not@valid",
		},
		expError: `Add: validate: invalid label "Not@valid"`,
	}, {
		desc: "With invalid hash",
		issuer: &Issuer{
			Label: "Test",
			Hash:  "SHA255",
		},
		expError: `Add: validate: invalid algorithm "SHA255"`,
	}, {
		desc: "With valid label",
		issuer: &Issuer{
			Label:  "Test",
			Hash:   HashSHA1,
			Secret: "x",
		},
		expConfig: "[gotp]\nprivate_key =\n\n[gotp \"issuer\"]\ntest = SHA1:x:6:30:\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err = cli.Add(c.issuer)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		got, err := os.ReadFile(cli.cfg.file)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, cli.cfg.file, c.expConfig, string(got))
	}
}
