// yt-mango: YT video metadata archiving utility
// Copyright (C) 2018  terorie

package main

import (
	"fmt"
	"os"

	"github.com/terorie/yt-mango/cmd"
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
