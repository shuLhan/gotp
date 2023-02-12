// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

module git.sr.ht/~shulhan/gotp

go 1.18

require (
	github.com/shuLhan/share v0.43.0
	golang.org/x/crypto v0.6.0
	golang.org/x/term v0.5.0
)

require golang.org/x/sys v0.5.0 // indirect

//replace github.com/shuLhan/share => ../share
