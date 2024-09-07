// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

module git.sr.ht/~shulhan/gotp

go 1.21

require git.sr.ht/~shulhan/pakakeh.go v0.57.0

require (
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/term v0.24.0 // indirect
)

//replace git.sr.ht/~shulhan/pakakeh.go => ../pakakeh.go
