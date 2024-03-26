// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

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
		desc:      `With nil issuer`,
		expConfig: ``,
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
		cfg = &config{
			dir: t.TempDir(),
		}

		rawConfig []byte
	)

	rawConfig = tdata.Input[`config.ini`]

	err = cfg.UnmarshalText(rawConfig)
	if err != nil {
		t.Fatal(err)
	}
	cli.cfg = cfg

	// Set the private key generated from openssl command.

	err = cli.SetPrivateKey(tdata.Flag[`private_key_openssl`])
	if err != nil {
		t.Fatal(err)
	}

	// Change the private key generated from ssh-keygen command.

	err = cli.SetPrivateKey(tdata.Flag[`private_key_openssh`])
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

	err = cli.cfg.loadPrivateKey()
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

func TestCli_ViewEncrypted(t *testing.T) {
	var (
		configDir = t.TempDir()

		cli *Cli
		err error
	)

	cli, err = NewCli(configDir)
	if err != nil {
		t.Fatal(err)
	}

	var privateKeyFile = filepath.Join(`testdata`, `keys`, `rsa-openssh.pem`)

	err = cli.SetPrivateKey(privateKeyFile)
	if err != nil {
		t.Fatal(err)
	}

	var issA *Issuer

	issA, err = NewIssuer(`testA`, `SHA1:TESTA`, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = cli.Add(issA)
	if err != nil {
		t.Fatal(err)
	}

	var gotIssA *Issuer

	gotIssA, err = cli.Get(`testA`)
	if err != nil {
		t.Fatal(err)
	}

	// Reset the raw issuer value for comparison.
	issA.raw = nil
	test.Assert(t, `Get: testA`, issA, gotIssA)
}

func TestCli_withPassphrase(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/cli_with_passphrase_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	// Prepare directory with private key.

	var (
		dirConfig      = t.TempDir()
		filePrivateKey = filepath.Join(dirConfig, privateKeyFile)
	)

	err = os.WriteFile(filePrivateKey, tdata.Input[`gotp.key`], 0600)
	if err != nil {
		t.Fatal(err)
	}

	var cli *Cli

	cli, err = NewCli(dirConfig)
	if err != nil {
		t.Fatal(err)
	}

	t.Run(`Add`, func(t *testing.T) {
		testAddWithPassphrase(t, tdata, cli)
	})

	t.Run(`Generate`, func(t *testing.T) {
		testGenerateWithPassphrase(t, tdata, cli)
	})

	t.Run(`Get`, func(t *testing.T) {
		testGetWithPassphrase(t, tdata, cli)
	})

	t.Run(`List`, func(t *testing.T) {
		var (
			expListLabel = []string{`test-sha1`, `test-sha256`, `test-sha512`}
			gotListLabel = cli.List()
		)
		test.Assert(t, `List`, expListLabel, gotListLabel)
	})

	t.Run(`Remove`, func(t *testing.T) {
		testRemoveWithPassphrase(t, tdata, cli)
	})

	t.Run(`Rename`, func(t *testing.T) {
		testRenameWithPassphrase(t, tdata, cli)
	})

	t.Run(`RemovePrivateKey`, func(t *testing.T) {
		testRemovePrivateKeyWithPassphrase(t, tdata, cli)
	})

	t.Run(`SetPrivateKey`, func(t *testing.T) {
		testSetPrivateKeyWithPassphrase(t, tdata, cli)
	})
}

func testAddWithPassphrase(t *testing.T, tdata *test.Data, cli *Cli) {
	var (
		pass  = string(tdata.Input[`gotp.pass`]) + "\r\n"
		lines = bytes.Split(tdata.Input[`list_raw_issuer`], []byte{'\n'})

		line        []byte
		labelIssuer [][]byte
		issuer      *Issuer
		err         error
	)

	for _, line = range lines {
		labelIssuer = bytes.Split(line, []byte{'='})

		issuer, err = NewIssuer(string(labelIssuer[0]), string(labelIssuer[1]), nil)
		if err != nil {
			t.Fatal(err)
		}

		mockTermrw.BufRead.WriteString(pass)

		err = cli.Add(issuer)
		if err != nil {
			t.Fatal(err)
		}
	}

	assertGotpConf(t, cli, string(tdata.Output[`gotp.conf:encrypted`]))

	mockTermrw.BufRead.Reset()
}

func testGenerateWithPassphrase(t *testing.T, tdata *test.Data, cli *Cli) {
	type testCase struct {
		label      string
		pass       string
		expListOTP []string
		n          int
	}

	var validPass = string(tdata.Input[`gotp.pass`]) + "\r\n"

	var listCase = []testCase{{
		label:      `test-sha1`,
		n:          3,
		pass:       validPass,
		expListOTP: []string{`002561`, `439480`, `508390`},
	}, {
		label:      `test-sha256`,
		n:          3,
		pass:       validPass,
		expListOTP: []string{`182691`, `322218`, `699844`},
	}, {
		label:      `test-sha512`,
		n:          3,
		pass:       validPass,
		expListOTP: []string{`595992`, `757602`, `224726`},
	}}

	var (
		c          testCase
		gotListOTP []string
		err        error
	)

	for _, c = range listCase {
		mockTermrw.BufRead.WriteString(c.pass)

		gotListOTP, err = cli.Generate(c.label, c.n)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.label, c.expListOTP, gotListOTP)
	}
	mockTermrw.BufRead.Reset()
}

func testGetWithPassphrase(t *testing.T, tdata *test.Data, cli *Cli) {
	type testCase struct {
		label     string
		expIssuer string
	}

	var listCase = []testCase{{
		label:     `test-sha1`,
		expIssuer: string(tdata.Output[`get:test-sha1`]),
	}, {
		label:     `test-sha256`,
		expIssuer: string(tdata.Output[`get:test-sha256`]),
	}, {
		label:     `test-sha512`,
		expIssuer: string(tdata.Output[`get:test-sha512`]),
	}}

	var (
		pass = string(tdata.Input[`gotp.pass`]) + "\r\n"

		c      testCase
		issuer *Issuer
		err    error
	)

	for _, c = range listCase {
		mockTermrw.BufRead.WriteString(pass)

		issuer, err = cli.Get(c.label)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, c.label, c.expIssuer, issuer.String())
	}
	mockTermrw.BufRead.Reset()
}

func testRemoveWithPassphrase(t *testing.T, tdata *test.Data, cli *Cli) {
	var pass = string(tdata.Input[`gotp.pass`]) + "\r\n"
	mockTermrw.BufRead.WriteString(pass)

	var err = cli.Remove(`test-sha512`)
	if err != nil {
		t.Fatal(err)
	}

	assertGotpConf(t, cli, string(tdata.Output[`gotp.conf:remove:encrypted`]))

	mockTermrw.BufRead.Reset()
}

// The Rename method does not require private key.
func testRenameWithPassphrase(t *testing.T, tdata *test.Data, cli *Cli) {
	var pass = string(tdata.Input[`gotp.pass`]) + "\r\n"
	mockTermrw.BufRead.WriteString(pass)

	var err = cli.Rename(`test-sha1`, `renamed-sha1`)
	if err != nil {
		t.Fatal(err)
	}

	assertGotpConf(t, cli, string(tdata.Output[`gotp.conf:rename:encrypted`]))

	mockTermrw.BufRead.Reset()
}

func testRemovePrivateKeyWithPassphrase(t *testing.T, tdata *test.Data, cli *Cli) {
	var pass = string(tdata.Input[`gotp.pass`]) + "\r\n"

	mockTermrw.BufRead.WriteString(pass)

	var err = cli.RemovePrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	assertGotpConf(t, cli, string(tdata.Output[`gotp.conf`]))

	var fileGotpKey = filepath.Join(cli.cfg.dir, privateKeyFile)

	_, err = os.Stat(fileGotpKey)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf(`expecting gotp.key to be removed, but still exists`)
	}

	mockTermrw.BufRead.Reset()
}

func testSetPrivateKeyWithPassphrase(t *testing.T, tdata *test.Data, cli *Cli) {
	// Write the private key file.
	var newPrivateKeyFile = filepath.Join(cli.cfg.dir, `new.key`)

	var err = os.WriteFile(newPrivateKeyFile, tdata.Input[`gotp.key`], 0600)
	if err != nil {
		t.Fatal(err)
	}

	var pass = string(tdata.Input[`gotp.pass`]) + "\r\n"
	mockTermrw.BufRead.WriteString(pass)

	err = cli.SetPrivateKey(newPrivateKeyFile)
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(cli.cfg.privateKeyFile)
	if err != nil {
		t.Fatal(err)
	}

	// When SetPrivateKey success, the output is somehow not
	// deterministic.
	// So we need to compare them with two possible outcomes.

	var gotpConf = filepath.Join(cli.cfg.dir, configFile)
	var rawconf []byte

	rawconf, err = os.ReadFile(gotpConf)
	if err != nil {
		t.Fatal(err)
	}

	rawconf = bytes.TrimSpace(rawconf)

	var gotConf = string(rawconf)
	var exp = string(tdata.Output[`gotp.conf:set-private-key:encrypted`])

	if exp != gotConf {
		exp = string(tdata.Output[`gotp.conf:set-private-key:encrypted:alt`])
		test.Assert(t, `assertGotpConf`, exp, string(rawconf))
	}

	mockTermrw.BufRead.Reset()
}

func assertGotpConf(t *testing.T, cli *Cli, exp string) {
	var (
		gotpConf = filepath.Join(cli.cfg.dir, configFile)

		rawconf []byte
		err     error
	)

	rawconf, err = os.ReadFile(gotpConf)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, `assertGotpConf`, exp, string(rawconf))
}
