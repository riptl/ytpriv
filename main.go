// ytwrk: YT video metadata archiving utility
// Copyright (C) 2019  terorie

package main

import (
	"fmt"
	"os"

	"github.com/terorie/ytwrk/cmd"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(cmd.Version)
		os.Exit(0)
	}

	if err := cmd.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
