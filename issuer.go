// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/ini"
	"github.com/shuLhan/share/lib/totp"
)

// Issuer contains the configuration for single TOTP issuer, including
// their unique label, algorithm, secret key, and number of digits.
type Issuer struct {
	Name     string
	Label    string
	Hash     string
	Secret   string // The secret value in base32.
	raw      []byte
	Digits   int
	TimeStep int
}

// NewIssuer create and initialize new issuer from raw value.
// If the rsaPrivateKey is not nil, that means the rawConfig is encrypted.
func NewIssuer(label, rawConfig string, rsaPrivateKey *rsa.PrivateKey) (issuer *Issuer, err error) {
	var (
		logp = `NewIssuer`

		vals   []string
		vbytes []byte
	)

	if rsaPrivateKey != nil {
		vbytes, err = base64.StdEncoding.DecodeString(rawConfig)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		vbytes, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaPrivateKey, vbytes, nil)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		rawConfig = string(vbytes)
	}

	vals = strings.Split(rawConfig, valueSeparator)
	if len(vals) < 2 {
		return nil, fmt.Errorf(`%s: invalid value %q`, logp, rawConfig)
	}
	issuer = &Issuer{
		Label:  label,
		Hash:   vals[0],
		Secret: vals[1],
	}
	if len(vals) >= 3 {
		issuer.Digits, err = strconv.Atoi(vals[2])
		if err != nil {
			return nil, fmt.Errorf(`%s: invalid digits %s: %w`, logp, vals[2], err)
		}
	} else {
		issuer.Digits = totp.DefCodeDigits
	}
	if len(vals) >= 4 {
		issuer.TimeStep, err = strconv.Atoi(vals[3])
		if err != nil {
			return nil, fmt.Errorf(`%s: invalid time step %s: %w`, logp, vals[3], err)
		}
	} else {
		issuer.TimeStep = totp.DefTimeStep
	}
	if len(vals) >= 5 {
		issuer.Name = vals[4]
	}

	return issuer, nil
}

func (issuer *Issuer) String() string {
	return fmt.Sprintf(`%s:%s:%d:%d:%s`, issuer.Hash, issuer.Secret,
		issuer.Digits, issuer.TimeStep, issuer.Name)
}

// pack the Issuer into string separated by `:`.
// If the privateKey is not nil, the string will be encrypted and encoded to
// base64.
func (issuer *Issuer) pack(privateKey *rsa.PrivateKey) (value string, err error) {
	var (
		logp      = `pack`
		plainText = issuer.String()
		rng       = rand.Reader
	)

	issuer.raw = []byte(plainText)
	if privateKey == nil {
		return string(issuer.raw), nil
	}

	issuer.raw, err = rsa.EncryptOAEP(sha256.New(), rng, &privateKey.PublicKey, issuer.raw, nil)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	value = base64.StdEncoding.EncodeToString(issuer.raw)

	return value, nil
}

func (issuer *Issuer) validate() (err error) {
	var (
		logp = `validate`
	)

	if !ini.IsValidVarName(issuer.Label) {
		return fmt.Errorf(`%s: invalid label %q`, logp, issuer.Label)
	}
	issuer.Hash = strings.ToUpper(issuer.Hash)
	switch issuer.Hash {
	case ``:
		issuer.Hash = defaultHash
	case HashSHA1, HashSHA256, HashSHA512:
		// NOOP
	default:
		return fmt.Errorf(`%s: invalid algorithm %q`, logp, issuer.Hash)
	}

	if len(issuer.Secret) == 0 {
		return fmt.Errorf(`%s: empty key`, logp)
	}
	if issuer.Digits <= 0 {
		issuer.Digits = totp.DefCodeDigits
	}
	if issuer.TimeStep <= 0 {
		issuer.TimeStep = totp.DefTimeStep
	}

	return nil
}
