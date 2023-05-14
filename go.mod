// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

module git.sr.ht/~shulhan/gotp

go 1.19

require (
	github.com/shuLhan/share v0.46.0
	golang.org/x/crypto v0.9.0
	golang.org/x/term v0.8.0
)

require golang.org/x/sys v0.8.0 // indirect

//replace github.com/shuLhan/share => ../share
