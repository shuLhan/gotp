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
)

const (
	valueSeparator = ":"
)

type config struct {
	PrivateKey string            `ini:"gotp::private_key"`
	Issuers    map[string]string `ini:"gotp:issuer"`

	file       string
	isNotExist bool
	privateKey *rsa.PrivateKey // Only RSA private key can do encryption.
}

func newConfig(file string) (cfg *config, err error) {
	logp := "newConfig"

	cfg = &config{
		Issuers: make(map[string]string),
		file:    file,
	}

	in, err := ini.Open(file)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%s: Open %q: %w", logp, file, err)
		}
		cfg.isNotExist = true
	}

	if cfg.isNotExist {
		dir := filepath.Dir(file)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return nil, fmt.Errorf("%s: MkdirAll %q: %w", logp, dir, err)
		}
		return cfg, nil
	}

	err = in.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return cfg, nil
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
		return fmt.Errorf("duplicate issuer name %q", issuer.Label)
	}

	value, err = issuer.pack(cfg.privateKey)
	if err != nil {
		return err
	}

	cfg.Issuers[issuer.Label] = value

	return nil
}

//
// get the issuer by its name.
//
func (cfg *config) get(name string) (issuer *Issuer, err error) {
	logp := "get"

	name = strings.ToLower(name)

	v, ok := cfg.Issuers[name]
	if !ok {
		return nil, fmt.Errorf("%s: issuer %q not found", logp, name)
	}

	issuer, err = NewIssuer(name, v, cfg.privateKey)
	if err != nil {
		return nil, fmt.Errorf("%s %q: %w", logp, name, err)
	}

	return issuer, nil
}

// save the config to file.
func (cfg *config) save() (err error) {
	logp := "save"

	b, err := ini.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("%s %s: %w", logp, cfg.file, err)
	}

	err = os.WriteFile(cfg.file, b, 0600)
	if err != nil {
		return fmt.Errorf("%s %s: %w", logp, cfg.file, err)
	}

	return nil
}
