package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
