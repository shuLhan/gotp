// SPDX-FileCopyrightText: 2024 Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

package gotp

import (
	"crypto/rand"
	"os"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test/mock"
)

// Mock the termrw for reading passphrase.
var mockTermrw = &mock.ReadWriter{}

// mockRandReader mock the reader for crypto [rand.Reader] which is used to
// provides predictable result in [libcrypto.EncryptOaep].
var mockRandReader = mock.NewRandReader([]byte(`gotptest`))

func TestMain(m *testing.M) {
	termrw = mockTermrw

	// Overwrite the random reader to provide predictable result.
	rand.Reader = mockRandReader

	// Overwrite current time for predictable OTP.
	timeNow = func() time.Time {
		return time.Date(2024, time.January, 30, 00, 03, 00, 00, time.UTC)
	}

	os.Exit(m.Run())
}
