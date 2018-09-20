package pretty

import (
	"github.com/mattn/go-isatty"
	"os"
	"strings"
)

var isTTY bool

func init() {
	term := os.Getenv("TERM")

	isTTY = strings.HasPrefix(term, "xterm") ||
		isatty.IsTerminal(os.Stdout.Fd())
}
