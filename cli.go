// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	_ "embed"
	"encoding/base32"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	libos "github.com/shuLhan/share/lib/os"
	"github.com/shuLhan/share/lib/totp"
)

// Readme embed the README.md, rendered in "gotp help".
//
//go:embed README.md
var Readme string

// Cli define the command line interface for gotp program.
type Cli struct {
	cfg *config
}

// NewCli create and initialize new CLI for gotp program.
func NewCli(configDir string) (cli *Cli, err error) {
	var (
		logp = `NewCli`

		cfgFile string
	)

	cli = &Cli{}

	cfgFile = filepath.Join(configDir, configFile)

	cli.cfg, err = newConfig(cfgFile)
	if err != nil {
		return nil, fmt.Errorf(`%s: UserConfigDir: %w`, logp, err)
	}

	return cli, nil
}

// Add new issuer to the config.
func (cli *Cli) Add(issuer *Issuer) (err error) {
	if issuer == nil {
		return nil
	}

	var logp = `Add`

	err = cli.cfg.loadPrivateKey()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = cli.add(issuer)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = cli.cfg.save()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// Generate n number of OTP from given issuer name.
func (cli *Cli) Generate(label string, n int) (listOtp []string, err error) {
	var (
		logp   = `Generate`
		b32Enc = base32.StdEncoding.WithPadding(base32.NoPadding)

		cryptoHash totp.CryptoHash
		issuer     *Issuer
		secret     []byte
		proto      totp.Protocol
	)

	err = cli.cfg.loadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	issuer, err = cli.cfg.get(label)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	secret, err = b32Enc.DecodeString(issuer.Secret)
	if err != nil {
		return nil, fmt.Errorf(`%s: secret is not a valid base32 encoding: %w`, logp, err)
	}

	switch issuer.Hash {
	case HashSHA256:
		cryptoHash = totp.CryptoHashSHA256
	case HashSHA512:
		cryptoHash = totp.CryptoHashSHA512
	default:
		cryptoHash = totp.CryptoHashSHA1
	}

	proto = totp.New(cryptoHash, issuer.Digits, issuer.TimeStep)

	listOtp, err = proto.GenerateN(secret, n)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return listOtp, nil
}

// Get the stored Issuer by its label.
func (cli *Cli) Get(label string) (issuer *Issuer, err error) {
	var logp = `Get`

	if cli.cfg.privateKey == nil {
		err = cli.cfg.loadPrivateKey()
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	issuer, err = cli.cfg.get(label)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return issuer, nil
}

// Import the TOTP configuration from file format based on provider.
func (cli *Cli) Import(providerName, file string) (n int, err error) {
	var (
		logp = `Import`

		issuers []*Issuer
		issuer  *Issuer
	)

	err = cli.cfg.loadPrivateKey()
	if err != nil {
		return 0, fmt.Errorf(`%s: %w`, logp, err)
	}

	providerName = strings.ToLower(providerName)
	switch providerName {
	case providerNameAegis:
	default:
		return 0, fmt.Errorf(`%s: unknown provider %q`, logp, providerName)
	}

	issuers, err = parseProviderAegis(file)
	if err != nil {
		return 0, fmt.Errorf(`%s: %w`, logp, err)
	}

	for _, issuer = range issuers {
		err = issuer.validate()
		if err != nil {
			return 0, fmt.Errorf(`%s: %w`, logp, err)
		}

		err = cli.cfg.add(issuer)
		if err != nil {
			return 0, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	err = cli.cfg.save()
	if err != nil {
		return 0, fmt.Errorf(`%s: %w`, logp, err)
	}

	return len(issuers), nil
}

// List all labels sorted in ascending order.
func (cli *Cli) List() (labels []string) {
	var (
		label string
	)

	for label = range cli.cfg.Issuers {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels
}

// Remove a TOTP configuration by its label.
func (cli *Cli) Remove(label string) (err error) {
	var (
		logp = `Remove`

		ok bool
	)

	label = strings.ToLower(label)

	_, ok = cli.cfg.Issuers[label]
	if !ok {
		return fmt.Errorf(`%s: %q not exist`, logp, label)
	}

	delete(cli.cfg.Issuers, label)

	err = cli.cfg.save()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// RemovePrivateKey decrypt the issuer's value (hash:secret...) using
// current private key and store it back to file as plain text.
// The current private key file will be removed from gotp directory.
//
// If no private key file, this method does nothing.
func (cli *Cli) RemovePrivateKey() (err error) {
	var logp = `RemovePrivateKey`

	err = cli.cfg.loadPrivateKey()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	if cli.cfg.privateKey == nil {
		// The private key file is not exist.
		return nil
	}

	var (
		oldPrivateKey = cli.cfg.privateKey
		oldIssuers    = cli.cfg.Issuers

		issuer *Issuer
		label  string
		raw    string
	)

	cli.cfg.privateKey = nil
	cli.cfg.Issuers = map[string]string{}

	for label, raw = range oldIssuers {
		// Decrypt the issuer using old private key.
		issuer, err = NewIssuer(label, raw, oldPrivateKey)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}

		// Add it to the config back as plain text.
		err = cli.cfg.add(issuer)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	err = cli.cfg.save()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = os.Remove(cli.cfg.privateKeyFile)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	cli.cfg.privateKeyFile = ``

	return nil
}

// Rename a label to newLabel.
// It will return an error if the label parameter is not exist or newLabel
// already exist.
func (cli *Cli) Rename(label, newLabel string) (err error) {
	var (
		logp = `Rename`

		rawValue string
		ok       bool
	)

	label = strings.ToLower(label)
	rawValue, ok = cli.cfg.Issuers[label]
	if !ok {
		return fmt.Errorf(`%s: %q not exist`, logp, label)
	}

	newLabel = strings.ToLower(newLabel)
	_, ok = cli.cfg.Issuers[newLabel]
	if ok {
		return fmt.Errorf(`%s: new label %q already exist`, logp, newLabel)
	}

	delete(cli.cfg.Issuers, label)

	cli.cfg.Issuers[newLabel] = rawValue

	err = cli.cfg.save()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// SetPrivateKey encrypt all the OTP secret using the new private key.
// The only accepted private key is RSA.
// If the pkeyFile is valid, it will be copied to
// "$XDG_CONFIG_DIR/gotp/gotp.key".
func (cli *Cli) SetPrivateKey(pkeyFile string) (err error) {
	var (
		logp          = `SetPrivateKey`
		oldIssuers    = cli.cfg.Issuers
		oldPrivateKey = cli.cfg.privateKey
	)

	cli.cfg.privateKeyFile = pkeyFile

	err = cli.cfg.loadPrivateKey()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		issuer *Issuer
		label  string
		raw    string
	)

	cli.cfg.Issuers = map[string]string{}

	for label, raw = range oldIssuers {
		// Decrypt the old issuer using old private key.
		issuer, err = NewIssuer(label, raw, oldPrivateKey)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}

		// Add it to the config back using new private key.
		err = cli.cfg.add(issuer)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	err = cli.cfg.save()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var expPrivateKeyPath = filepath.Join(cli.cfg.dir, privateKeyFile)

	if expPrivateKeyPath != pkeyFile {
		// Copy the private key file if the path is not
		// "$configDir/gotp.key".
		err = libos.Copy(expPrivateKeyPath, pkeyFile)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
		cli.cfg.privateKeyFile = expPrivateKeyPath
	}

	return nil
}

func (cli *Cli) add(issuer *Issuer) (err error) {
	err = issuer.validate()
	if err != nil {
		return err
	}
	err = cli.cfg.add(issuer)
	if err != nil {
		return err
	}
	return nil
}
