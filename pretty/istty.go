package pretty

import (
		"os"
	"strings"
	"github.com/mattn/go-isatty"
)

var isTTY bool

func init() {
	term := os.Getenv("TERM")

	isTTY = strings.HasPrefix(term, "xterm") ||
		isatty.IsTerminal(os.Stdout.Fd())
}
