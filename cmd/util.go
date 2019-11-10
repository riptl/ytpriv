package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func cmdFunc(f func(*cobra.Command, []string)error) func(*cobra.Command, []string) {
	return func(c *cobra.Command, args []string) {
		err := f(c, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func stdinOrArgs(out chan<- string, args []string) {
	defer close(out)
	if len(args) == 0 {
		scn := bufio.NewScanner(os.Stdin)
		for scn.Scan() {
			out <- scn.Text()
		}
	} else {
		for _, item := range args {
			out <- item
		}
	}
}
