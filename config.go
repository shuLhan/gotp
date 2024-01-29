// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"crypto"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	libcrypto "github.com/shuLhan/share/lib/crypto"
	"github.com/shuLhan/share/lib/ini"
)

const (
	valueSeparator = `:`
)

type config struct {
	privateKey *rsa.PrivateKey // Only RSA private key can do encryption.

	Issuers map[string]string `ini:"gotp:issuer"`

	dir            string
	file           string
	privateKeyFile string
}

func newConfig(file string) (cfg *config, err error) {
	var logp = `newConfig`

	cfg = &config{
		dir:  filepath.Dir(file),
		file: file,
	}

	cfg.privateKeyFile = filepath.Join(cfg.dir, privateKeyFile)

	var content []byte

	content, err = os.ReadFile(file)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf(`%s: Open %q: %w`, logp, file, err)
		}

		err = os.MkdirAll(cfg.dir, 0700)
		if err != nil {
			return nil, fmt.Errorf(`%s: MkdirAll %q: %w`, logp, cfg.dir, err)
		}
	}

	err = cfg.UnmarshalText(content)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return cfg, nil
}

// UnmarshalText load configuration from raw bytes.
func (cfg *config) UnmarshalText(content []byte) (err error) {
	var logp = `UnmarshalText`

	cfg.Issuers = make(map[string]string)

	if len(content) > 0 {
		err = ini.Unmarshal(content, cfg)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	return nil
}

// MarshalText convert the config object back to INI format.
func (cfg *config) MarshalText() (text []byte, err error) {
	var logp = `MarshalText`

	text, err = ini.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	return text, nil
}

func (cfg *config) add(issuer *Issuer) (err error) {
	var (
		value string
		exist bool
	)

	if issuer == nil {
		return nil
	}

	_, exist = cfg.Issuers[issuer.Label]
	if exist {
		return fmt.Errorf(`duplicate issuer name %q`, issuer.Label)
	}

	value, err = issuer.pack(cfg.privateKey)
	if err != nil {
		return err
	}

	cfg.Issuers[issuer.Label] = value

	return nil
}

// get the issuer by its name.
func (cfg *config) get(name string) (issuer *Issuer, err error) {
	var (
		logp = `get`

		v  string
		ok bool
	)

	name = strings.ToLower(name)

	v, ok = cfg.Issuers[name]
	if !ok {
		return nil, fmt.Errorf(`%s: issuer %q not found`, logp, name)
	}

	issuer, err = NewIssuer(name, v, cfg.privateKey)
	if err != nil {
		return nil, fmt.Errorf(`%s %q: %w`, logp, name, err)
	}

	return issuer, nil
}

// loadPrivateKey parse the RSA private key with optional passphrase.
// It will return nil if private key file does not exist.
func (cfg *config) loadPrivateKey() (err error) {
	var logp = `loadPrivateKey`

	_, err = os.Stat(cfg.privateKeyFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var privateKey crypto.PrivateKey

	privateKey, err = libcrypto.LoadPrivateKeyInteractive(termrw, cfg.privateKeyFile)
	if err != nil {
		return fmt.Errorf(`%s %q: %w`, logp, cfg.privateKeyFile, err)
	}

	var ok bool

	cfg.privateKey, ok = privateKey.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf(`%s: invalid or unsupported private key`, logp)
	}

	return nil
}

// save the config to file.
func (cfg *config) save() (err error) {
	if len(cfg.file) == 0 {
		return nil
	}

	var (
		logp = `save`

		b []byte
	)

	b, err = cfg.MarshalText()
	if err != nil {
		return fmt.Errorf(`%s %s: %w`, logp, cfg.file, err)
	}

	err = os.WriteFile(cfg.file, b, 0600)
	if err != nil {
		return fmt.Errorf(`%s %s: %w`, logp, cfg.file, err)
	}

	return nil
}
