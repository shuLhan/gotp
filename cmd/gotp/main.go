// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"git.sr.ht/~shulhan/gotp"
)

const (
	cmdName     = "gotp"
	cmdAdd      = "add"
	cmdGenerate = "gen"
	cmdImport   = "import"
	cmdList     = "list"
	cmdRemove   = "remove"
	cmdRename   = "rename"
)

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Println(gotp.Readme)
		os.Exit(2)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
	}

	cmd := strings.ToLower(args[0])

	switch cmd {
	case cmdAdd:
		if len(args) < 3 {
			log.Printf("%s %s: missing parameters", cmdName, cmd)
			os.Exit(1)
		}
	case cmdGenerate:
		if len(args) < 2 {
			log.Printf("%s %s: missing parameters", cmdName, cmd)
			os.Exit(1)
		}
	case cmdImport:
		if len(args) <= 2 {
			log.Printf("%s %s: missing parameters", cmdName, cmd)
			os.Exit(1)
		}
	case cmdList:
		// NOOP.
	case cmdRemove:
		if len(args) <= 1 {
			log.Printf("%s %s: missing parameters", cmdName, cmd)
			os.Exit(1)
		}
	case cmdRename:
		if len(args) <= 2 {
			log.Printf("%s %s: missing parameters", cmdName, cmd)
			os.Exit(1)
		}
	default:
		log.Printf("%s: unknown command %q", cmdName, cmd)
		flag.Usage()
	}

	cli, err := gotp.NewCli()
	if err != nil {
		log.Printf("%s: %s", cmdName, err)
		os.Exit(1)
	}

	switch cmd {
	case cmdAdd:
		doAdd(cli, args)
	case cmdGenerate:
		doGenerate(cli, args)
	case cmdImport:
		doImport(cli, args)
	case cmdList:
		doList(cli)
	case cmdRemove:
		doRemove(cli, args)
	case cmdRename:
		doRename(cli, args)
	}
}

func doAdd(cli *gotp.Cli, args []string) {
	label := args[1]
	rawConfig := args[2]
	issuer, err := gotp.NewIssuer(label, rawConfig, nil)
	if err != nil {
		log.Printf("%s: %s", cmdName, err)
		os.Exit(1)
	}
	err = cli.Add(issuer)
	if err != nil {
		log.Printf("%s: %s", cmdName, err)
		os.Exit(1)
	}
	fmt.Println("OK")
}

func doGenerate(cli *gotp.Cli, args []string) {
	var (
		label     = args[1]
		n     int = 1
		err   error
	)
	if len(args) >= 3 {
		n, err = strconv.Atoi(args[2])
		if err != nil {
			log.Printf("%s: %s", cmdName, err)
			os.Exit(1)
		}
	}
	listOtp, err := cli.Generate(label, n)
	if err != nil {
		log.Printf("%s: %s", cmdName, err)
		os.Exit(1)
	}
	for _, otp := range listOtp {
		fmt.Println(otp)
	}
}

func doImport(cli *gotp.Cli, args []string) {
	var (
		providerName = args[1]
		file         = args[2]
	)
	n, err := cli.Import(providerName, file)
	if err != nil {
		log.Printf("%s: %s", cmdName, err)
		os.Exit(1)
	}
	fmt.Printf("OK - %d imported", n)
}

func doList(cli *gotp.Cli) {
	labels := cli.List()
	for _, label := range labels {
		fmt.Println(label)
	}
}

func doRemove(cli *gotp.Cli, args []string) {
	label := args[1]
	err := cli.Remove(label)
	if err != nil {
		log.Printf("%s: %s", cmdName, err)
		os.Exit(1)
	}
	fmt.Println("OK")
}

func doRename(cli *gotp.Cli, args []string) {
	label := args[1]
	newLabel := args[2]

	err := cli.Rename(label, newLabel)
	if err != nil {
		log.Printf("%s: %s", cmdName, err)
		os.Exit(1)
	}
	fmt.Println("OK")
}
