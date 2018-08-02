// yt-mango: YT video metadata archiving utility
// Copyright (C) 2018  terorie

package main

import (
	"fmt"
	"os"
	"log"
	"github.com/terorie/yt-mango/cmd"
)

func main() {
	// All diagnostics (logging) should go to stderr
	log.SetOutput(os.Stderr)

	if err := cmd.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
