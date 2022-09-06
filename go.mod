// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

module git.sr.ht/~shulhan/gotp

go 1.18

require (
	github.com/shuLhan/share v0.41.0
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90
	golang.org/x/term v0.0.0-20220722155259-a9ba230a4035
)

require golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect

//replace github.com/shuLhan/share => ../share
