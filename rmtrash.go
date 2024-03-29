// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

//go:build darwin
// +build darwin

package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

var (
	version     = "0.0.3"
	flagVersion = flag.Bool("v", false, "print out version info")
	flagUser    = flag.String("u", "", "move the file to some other user's trash")
	_           = flag.Bool("r", false, "no effect, for compatibility with rm -r")
	flagF       = flag.Bool("f", false, "no effect, for compatibility with rm -f")
	flagRF      = flag.Bool("rf", false, "no effect, for compatibility with rm -rf")
	flagFR      = flag.Bool("fr", false, "no effect, for compatibility with rm -fr")
)

func usage() {
	fmt.Fprintf(os.Stderr, `usage: rmtrash [-u USERNAME] FILENAME
options:
`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(os.Args) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	if *flagVersion {
		fmt.Fprintf(os.Stdout, `rmtrash %s

Source: https://changkun.de/s/rmtrash
`, version)
		return
	}

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(2)
	}

	u, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot find the current user: %v", err)
		os.Exit(1)
	}
	uname := u.Username
	if *flagUser != "" {
		uname = *flagUser

		// Early checking if the given user exists
		if _, err := os.Stat(fmt.Sprintf("/Users/%s", uname)); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "User \"%s\" does not exists\n", uname)
			os.Exit(1)
		}
	}

	// If an argument contains a wildcard, such as rmtrash -rf ./x/*
	// and the matches is ./x/x1 ./x/x2 ./x/x3, then the flag.Args will
	// return all matches automatically. Process them all. Issue #2.
	for _, in := range flag.Args() {
		move(in, uname)
	}
}

func move(in, uname string) {
	src, err := filepath.Abs(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid filepath: %s", in)
		os.Exit(1)
	}

	_, err = os.Lstat(src)
	if os.IsNotExist(err) {
		if *flagF || *flagRF || *flagFR {
			return
		}
		fmt.Fprintf(os.Stderr, "no such file or directory\n")
		os.Exit(1)
	}

	// Try always append time info, full disk access is required to
	// check whether the file is existed or not.
	_, file := filepath.Split(src)
	fext := filepath.Ext(file)
	fname := file[0 : len(file)-len(fext)]
	dst := fmt.Sprintf("/Users/%s/.Trash/%s.%s", uname,
		fname, time.Now().Format("20060102150405"))
	if fext != "" {
		dst = fmt.Sprintf("%s%s", dst, fext)
	}

	// Move
	err = os.Rename(src, dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, `Could not move "%s" to the trash:
(Perhaps you don't have sufficient privileges?)
`, in)
		os.Exit(1)
	}
}
