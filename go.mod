// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

module git.sr.ht/~shulhan/gotp

go 1.23.4

require git.sr.ht/~shulhan/pakakeh.go v0.60.0

require (
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/exp v0.0.0-20250128182459-e0ece0dbea4c // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
)

//replace git.sr.ht/~shulhan/pakakeh.go => ../pakakeh.go
