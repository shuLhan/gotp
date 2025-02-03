// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: GPL-3.0-or-later

// Command gotp a command line interface to manage and generate Time-based One
// Time Password (TOTP).
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.sr.ht/~shulhan/gotp"
)

const (
	cmdName             = `gotp`
	cmdAdd              = `add`
	cmdExport           = `export`
	cmdGenerate         = `gen`
	cmdGet              = `get`
	cmdImport           = `import`
	cmdList             = `list`
	cmdRemove           = `remove`
	cmdRemovePrivateKey = `remove-private-key`
	cmdRename           = `rename`
	cmdSetPrivateKey    = `set-private-key`
	cmdVersion          = `version`
)

// defConfigDir default directory name for configuration.
const defConfigDir = `gotp`

func main() {
	var (
		cmd  string
		cli  *gotp.Cli
		err  error
		args []string
	)

	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Println(gotp.Readme)
		os.Exit(2)
	}
	flag.Parse()

	args = flag.Args()

	if len(args) == 0 {
		flag.Usage()
	}

	cmd = strings.ToLower(args[0])

	switch cmd {
	case cmdAdd:
		if len(args) < 3 {
			log.Printf(`%s %s: missing parameters`, cmdName, cmd)
			os.Exit(1)
		}
	case cmdExport:
		if len(args) < 2 {
			log.Fatalf(`%s %s: missing parameter: format`, cmdName, cmd)
		}

	case cmdGenerate:
		if len(args) < 2 {
			log.Printf(`%s %s: missing parameters`, cmdName, cmd)
			os.Exit(1)
		}
	case cmdGet:
		args = args[1:]
		if len(args) == 0 {
			log.Fatalf(`%s %s: missing parameters`, cmdName, cmd)
		}

	case cmdImport:
		if len(args) <= 2 {
			log.Printf(`%s %s: missing parameters`, cmdName, cmd)
			os.Exit(1)
		}
	case cmdList:
		// NOOP.
	case cmdRemove:
		if len(args) <= 1 {
			log.Printf(`%s %s: missing parameters`, cmdName, cmd)
			os.Exit(1)
		}
	case cmdRemovePrivateKey:
		// NOOP.
	case cmdRename:
		if len(args) <= 2 {
			log.Printf(`%s %s: missing parameters`, cmdName, cmd)
			os.Exit(1)
		}

	case cmdSetPrivateKey:
		if len(args) <= 1 {
			log.Printf(`%s %s: missing parameters`, cmdName, cmd)
			os.Exit(1)
		}

	case cmdVersion:
		fmt.Println(`gotp version`, gotp.Version)
		return

	default:
		log.Fatalf(`%s: unknown command %q`, cmdName, cmd)
	}

	var userConfigDir string

	userConfigDir, err = os.UserConfigDir()
	if err != nil {
		log.Fatalf(`%s: UserConfigDir: %s`, cmdName, err)
	}

	var configDir = filepath.Join(userConfigDir, defConfigDir)

	cli, err = gotp.NewCli(configDir)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
		os.Exit(1)
	}

	switch cmd {
	case cmdAdd:
		doAdd(cli, args)
	case cmdExport:
		doExport(cli, flag.Arg(1), flag.Arg(2))
	case cmdGenerate:
		doGenerate(cli, args)
	case cmdGet:
		doGet(cli, args[0])
	case cmdImport:
		doImport(cli, args)
	case cmdList:
		doList(cli)
	case cmdRemove:
		doRemove(cli, args)
	case cmdRemovePrivateKey:
		doRemovePrivateKey(cli)
	case cmdRename:
		doRename(cli, args)
	case cmdSetPrivateKey:
		doSetPrivateKey(cli, args)
	}
}

func doAdd(cli *gotp.Cli, args []string) {
	var (
		label     = args[1]
		rawConfig = args[2]

		issuer *gotp.Issuer
		err    error
	)

	issuer, err = gotp.NewIssuer(label, rawConfig, nil)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
		os.Exit(1)
	}
	err = cli.Add(issuer)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
		os.Exit(1)
	}
	fmt.Println(`OK`)
}

func doExport(cli *gotp.Cli, providerName string, exportFile string) {
	var (
		out *os.File
		err error
	)

	if len(exportFile) == 0 {
		out = os.Stdout
	} else {
		out, err = os.OpenFile(exportFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf(`export: %s`, err)
		}
	}

	err = cli.Export(out, providerName)
	if err != nil {
		goto out
	}

	return
out:
	log.Printf(`export: %s`, err)

	if len(exportFile) == 0 {
		err = out.Close()
		if err != nil {
			log.Printf(`export: %s`, err)
		}
	}
	os.Exit(1)
}

func doGenerate(cli *gotp.Cli, args []string) {
	var (
		label = args[1]
		n     = 1

		listOtp []string
		otp     string
		err     error
	)

	if len(args) >= 3 {
		n, err = strconv.Atoi(args[2])
		if err != nil {
			log.Printf(`%s: %s`, cmdName, err)
			os.Exit(1)
		}
	}

	listOtp, err = cli.Generate(label, n)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
		os.Exit(1)
	}
	for _, otp = range listOtp {
		fmt.Println(otp)
	}
}

// doGet execute the "get" command to print the issuer by "label".
func doGet(cli *gotp.Cli, label string) {
	var (
		issuer *gotp.Issuer
		err    error
	)

	issuer, err = cli.Get(label)
	if err != nil {
		log.Fatalf(`%s: %s`, cmdName, err)
	}

	fmt.Println(issuer.String())
}

func doImport(cli *gotp.Cli, args []string) {
	var (
		providerName = args[1]
		file         = args[2]

		n   int
		err error
	)
	n, err = cli.Import(providerName, file)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
		os.Exit(1)
	}
	fmt.Printf(`OK - %d imported`, n)
}

func doList(cli *gotp.Cli) {
	var (
		labels []string = cli.List()

		label string
	)

	for _, label = range labels {
		fmt.Println(label)
	}
}

func doRemove(cli *gotp.Cli, args []string) {
	var (
		label = args[1]

		err error
	)

	err = cli.Remove(label)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
		os.Exit(1)
	}
	fmt.Println(`OK`)
}

func doRemovePrivateKey(cli *gotp.Cli) {
	var err = cli.RemovePrivateKey()
	if err != nil {
		log.Fatalf(`%s: %s`, cmdName, err)
	}
}

func doRename(cli *gotp.Cli, args []string) {
	var (
		label    = args[1]
		newLabel = args[2]

		err error
	)

	err = cli.Rename(label, newLabel)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
		os.Exit(1)
	}
	fmt.Println(`OK`)
}

func doSetPrivateKey(cli *gotp.Cli, args []string) {
	var err = cli.SetPrivateKey(args[1])
	if err != nil {
		log.Fatalf(`%s: %s`, cmdName, err)
	}
}
