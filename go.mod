// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

module git.sr.ht/~shulhan/gotp

go 1.22.0

require git.sr.ht/~shulhan/pakakeh.go v0.58.0

require (
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/term v0.25.0 // indirect
)

//replace git.sr.ht/~shulhan/pakakeh.go => ../pakakeh.go
