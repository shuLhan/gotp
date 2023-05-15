// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"crypto/rsa"
	_ "embed"
	"encoding/base32"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shuLhan/share/lib/totp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

//go:embed README.md
var Readme string

type Cli struct {
	cfg *config
}

func NewCli() (cli *Cli, err error) {
	var (
		logp = `NewCli`

		userConfigDir string
		cfgFile       string
	)

	cli = &Cli{}

	userConfigDir, err = os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf(`%s: UserConfigDir: %w`, logp, err)
	}

	cfgFile = filepath.Join(userConfigDir, configDir, configFile)

	cli.cfg, err = newConfig(cfgFile)
	if err != nil {
		return nil, fmt.Errorf(`%s: UserConfigDir: %w`, logp, err)
	}

	if cli.cfg.isNotExist {
		cli.cfg.PrivateKey, err = cli.inputPrivateKey(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		if len(cli.cfg.PrivateKey) > 0 {
			cli.cfg.privateKey, err = loadPrivateKey(cli.cfg.PrivateKey, nil)
			if err != nil {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}
		}
		err = cli.cfg.save()
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	return cli, nil
}

func (cli *Cli) inputPrivateKey(stdin *os.File) (privateKeyFile string, err error) {
	fmt.Println(`Seems like this is your first time using this gotp.`)
	fmt.Println(`If you would like to encrypt the secret, please`)
	fmt.Printf(`enter the path to private key or enter to skip it: `)
	fmt.Fscanln(stdin, &privateKeyFile)

	return privateKeyFile, nil
}

// loadPrivateKey parse the RSA private key with optional passphrase.
func loadPrivateKey(privateKeyFile string, pass []byte) (rsaPrivateKey *rsa.PrivateKey, err error) {
	if len(privateKeyFile) == 0 {
		return nil, nil
	}

	var (
		logp           = `loadPrivateKey`
		errPassMissing = &ssh.PassphraseMissingError{}

		privateKey interface{}
		stdin      int
		rawPem     []byte
		ok         bool
	)

	rawPem, err = os.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	if len(pass) == 0 {
		privateKey, err = ssh.ParseRawPrivateKey(rawPem)
	} else {
		privateKey, err = ssh.ParseRawPrivateKeyWithPassphrase(rawPem, pass)
	}
	if err != nil {
		if !errors.As(err, &errPassMissing) {
			return nil, fmt.Errorf(`%s %q: %w`, logp, privateKeyFile, err)
		}

		fmt.Printf(`Enter passphrase for %s: `, privateKeyFile)

		stdin = int(os.Stdin.Fd())

		pass, err = term.ReadPassword(stdin)
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf(`%s %q: %w`, logp, privateKeyFile, err)
		}

		return loadPrivateKey(privateKeyFile, pass)
	}
	rsaPrivateKey, ok = privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf(`%s: invalid or unsupported private key`, logp)
	}

	return rsaPrivateKey, nil
}

// Add new issuer to the config.
func (cli *Cli) Add(issuer *Issuer) (err error) {
	if issuer == nil {
		return nil
	}

	var (
		logp = `Add`
	)

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

// Import the TOTP configuration from file format based on provider.
func (cli *Cli) Import(providerName, file string) (n int, err error) {
	var (
		logp = `Import`

		issuers []*Issuer
		issuer  *Issuer
	)

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
