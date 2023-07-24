// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewConfig(t *testing.T) {
	type testCase struct {
		expConfig  *config
		desc       string
		configFile string
		expError   string
	}

	var cases = []testCase{{
		desc:       `With file not exist`,
		configFile: `testdata/config-not-exist`,
		expConfig: &config{
			file:       `testdata/config-not-exist`,
			isNotExist: true,
		},
	}, {
		desc:       `With openssh rsa`,
		configFile: `testdata/with_private_key.conf`,
		expConfig: &config{
			PrivateKey: `testdata/keys/rsa-openssl.pem`,
			Issuers: map[string]string{
				`email-domain`: `XYZ`,
				`test`:         `ABCD`,
			},
			file: `testdata/with_private_key.conf`,
		},
	}}

	var (
		c         testCase
		gotConfig *config
		err       error
	)

	for _, c = range cases {
		t.Log(c.desc)

		gotConfig, err = newConfig(c.configFile)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}

		gotConfig.privateKey = nil

		test.Assert(t, `Issuer`, c.expConfig, gotConfig)
	}
}

func TestMarshaler(t *testing.T) {
	var (
		cfg = config{}

		tdata       *test.Data
		userHomeDir string
		err         error
	)

	tdata, err = test.LoadData(`testdata/config_marshaler_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.UnmarshalText(tdata.Input[`input.ini`])
	if err != nil {
		t.Fatal(err)
	}

	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	var expPrivateKey = filepath.Join(userHomeDir, `myprivatekey.pem`)

	test.Assert(t, `UnmarshalText: PrivateKey`, expPrivateKey, cfg.PrivateKey)

	var gotText []byte

	gotText, err = cfg.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, `MarshalText`, string(tdata.Output[`output.ini`]), string(gotText))
}
