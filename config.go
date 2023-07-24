// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/shuLhan/share/lib/ini"
	libos "github.com/shuLhan/share/lib/os"
)

const (
	valueSeparator = `:`
)

type config struct {
	privateKey *rsa.PrivateKey // Only RSA private key can do encryption.

	Issuers    map[string]string `ini:"gotp:issuer"`
	PrivateKey string            `ini:"gotp::private_key"`

	file       string
	isNotExist bool
}

func newConfig(file string) (cfg *config, err error) {
	var (
		logp = `newConfig`

		content    []byte
		isNotExist bool
	)

	content, err = os.ReadFile(file)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf(`%s: Open %q: %w`, logp, file, err)
		}

		var dir = filepath.Dir(file)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return nil, fmt.Errorf(`%s: MkdirAll %q: %w`, logp, dir, err)
		}
		isNotExist = true
	}

	cfg = &config{}

	err = cfg.UnmarshalText(content)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	cfg.isNotExist = isNotExist
	cfg.file = file

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

	if len(cfg.PrivateKey) != 0 {
		cfg.PrivateKey, err = libos.PathUnfold(cfg.PrivateKey)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	return nil
}

// MarshalText convert the config object back to INI format.
func (cfg *config) MarshalText() (text []byte, err error) {
	var logp = `MarshalText`

	if len(cfg.PrivateKey) != 0 {
		cfg.PrivateKey, err = libos.PathFold(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

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
