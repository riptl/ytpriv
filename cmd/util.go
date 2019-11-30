package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"

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

func signalContext(root context.Context) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	exitC := make(chan os.Signal)
	signal.Notify(exitC, os.Interrupt)
	go func() {
		select {
		case <-root.Done():
			break
		case <-exitC:
			cancel()
		}
	}()
	return ctx
}
