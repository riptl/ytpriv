// yt-mango: YT video metadata archiving utility
// Copyright (C) 2018  terorie

package main

import (
	"fmt"
	"os"
	"log"
	"github.com/terorie/yt-mango/cmd"
	"github.com/terorie/yt-mango/version"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version.Get())
		os.Exit(0)
	}

	// All diagnostics (logging) should go to stderr
	log.SetOutput(os.Stderr)

	if err := cmd.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
