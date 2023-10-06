// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestCli_inputPrivateKey(t *testing.T) {
	type testCase struct {
		desc       string
		privateKey string
		exp        string
	}

	var (
		cli = &Cli{
			cfg: &config{
				file:       `testdata/save.conf`,
				isNotExist: true,
			},
		}

		c                 testCase
		r                 *os.File
		w                 *os.File
		gotPrivateKeyFile string
		err               error
	)

	var cases = []testCase{{
		desc: `Without private key`,
		exp:  "[gotp]\nprivate_key =\n",
	}, {
		desc:       `With private key`,
		privateKey: `testdata/rsa`,
	}}

	for _, c = range cases {
		r, w, err = os.Pipe()
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprintf(w, "%s\n", c.privateKey)

		gotPrivateKeyFile, err = cli.inputPrivateKey(r)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, cli.cfg.file, c.privateKey, gotPrivateKeyFile)
	}
}

func TestCli_Add(t *testing.T) {
	type testCase struct {
		issuer    *Issuer
		desc      string
		expError  string
		expConfig string
	}

	var (
		cli = &Cli{
			cfg: &config{
				Issuers: make(map[string]string),
				file:    `testdata/add.conf`,
			},
		}

		err error
	)

	err = cli.cfg.save()
	if err != nil {
		t.Fatal(err)
	}

	var cases = []testCase{{
		desc: `With nil issuer`,
		expConfig: `[gotp]
private_key =
`,
	}, {
		desc: `With invalid label`,
		issuer: &Issuer{
			Label: `Not@valid`,
		},
		expError: `Add: validate: invalid label "Not@valid"`,
	}, {
		desc: `With invalid hash`,
		issuer: &Issuer{
			Label: `Test`,
			Hash:  `SHA255`,
		},
		expError: `Add: validate: invalid algorithm "SHA255"`,
	}, {
		desc: `With valid label`,
		issuer: &Issuer{
			Label:  `Test`,
			Hash:   HashSHA1,
			Secret: `x`,
		},
		expConfig: `[gotp "issuer"]
test = SHA1:x:6:30:

[gotp]
private_key =
`,
	}}

	var (
		c   testCase
		got []byte
	)

	for _, c = range cases {
		t.Log(c.desc)

		err = cli.Add(c.issuer)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}

		got, err = os.ReadFile(cli.cfg.file)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, cli.cfg.file, c.expConfig, string(got))
	}
}

func TestCli_SetPrivateKey(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/cli_SetPrivateKey_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		cli = &Cli{}
		cfg = &config{}

		rawConfig []byte
	)

	rawConfig = tdata.Input[`config.ini`]

	err = cfg.UnmarshalText(rawConfig)
	if err != nil {
		t.Fatal(err)
	}
	cli.cfg = cfg

	// Set the private key.

	err = cli.SetPrivateKey(tdata.Flag[`private_key_openssl`])
	if err != nil {
		t.Fatal(err)
	}

	// Change the private key.

	err = cli.SetPrivateKey(tdata.Flag[`private_key_openssl`])
	if err != nil {
		t.Fatal(err)
	}

	rawConfig, err = cli.cfg.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	// Load the encrypted raw config and compare the issuer.

	err = cfg.UnmarshalText(rawConfig)
	if err != nil {
		t.Fatal(err)
	}
	cli.cfg = cfg

	cli.cfg.privateKey, err = loadPrivateKey(cli.cfg.PrivateKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	var (
		gotLabels = cli.List()

		label  string
		issuer *Issuer
		got    bytes.Buffer
	)

	for _, label = range gotLabels {
		issuer, err = cli.cfg.get(label)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Fprintf(&got, "%s = %s\n", label, issuer.String())
	}

	test.Assert(t, `get all labels`, string(tdata.Output[`issuers`]), got.String())

	// Remove the private key, and compare the plain config.

	err = cli.RemovePrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	var gotConfig []byte

	gotConfig, err = cli.cfg.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	rawConfig = tdata.Input[`config.ini`]
	test.Assert(t, `RemovePrivateKey`, string(rawConfig), string(gotConfig))
}
