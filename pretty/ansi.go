package pretty

import (
	"bytes"
)

type Code string
type Codes []Code


type Effect interface {
	E(string) string
}

// Empty effect
type nilEffect struct{}
func (_ nilEffect) E(x string) string { return x }

// Custom effect
type customEffect func(string) string
func (e customEffect) E(x string) string { return e(x) }

const (
	RESET = Code("0")
	BOLD = Code("1")
	DIM = Code("2")
	ITALIC = Code("3")
	UNDERL = Code("4")
	INV = Code("7")
	HIDDEN = Code("8")
	STRIKE = Code("9")
	BLACK = Code("30")
	RED = Code("31")
	GREEN = Code("32")
	YELLOW = Code("33")
	BLUE = Code("34")
	MGNTA = Code("35")
	CYAN = Code("36")
	WHITE = Code("37")
	HBLACK = Code("90")
	HRED = Code("91")
	HGREEN = Code("92")
	HYELLOW = Code("93")
	HBLUE = Code("94")
	HMGNTA = Code("95")
	HCYAN = Code("96")
	HWHITE = Code("97")
)

func Add(x... Code) Codes {
	return Codes(x)
}

func (c Code) E(x string) string {
	if !isTTY { return x }
	return "\x1b[" + string(c) + "m" + x + "\x1b[0m"
}

func (cs Codes) E(x string) string {
	if !isTTY { return x }
	var b bytes.Buffer
	b.WriteString("\x1b[")
	for _, c := range cs {
		b.WriteRune(';')
		b.WriteString(string(c))
	}
	b.WriteRune('m')
	b.WriteString(x)
	b.WriteString("\x1b[0m")
	return b.String()
}

func Wrap(e Effect, wrapper string) Effect {
	if !isTTY { return nilEffect{} }
	return customEffect(func(s string) string {
		return e.E(wrapper[0:1]) + s + e.E(wrapper[1:2])
	})
}
