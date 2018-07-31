package util

import "bytes"

// Markdown escape map (ASCII)
// Inspired by https://github.com/golang-commonmark/markdown/blob/master/escape.go
type EscapeMap [2]uint64

func (e *EscapeMap) Get(index uint) bool {
	if index >= 128 { return false }
	high, low := index / 64, index % 64
	return e[high] & (1 << low) != 0
}

func (e *EscapeMap) Set(index uint, x bool) {
	if index >= 128 { return }
	high, low := index / 64, index % 64
	if x {
		e[high] = e[high] | 1 << low
	} else {
		e[high] = e[high] &^ 1 << low
	}
}

func (e EscapeMap) ToBuffer(src string, dest *bytes.Buffer) (err error) {
	for _, char := range src {
		if char < 0x80 && e.Get(uint(char)) {
			// Write backslash + char
			_, err = dest.Write([]byte{0x5c, byte(char)})
		} else {
			_, err = dest.WriteRune(char)
		}
	}
	return
}

func (e EscapeMap) ToString(src string) (string, error) {
	var buffer bytes.Buffer
	err := e.ToBuffer(src, &buffer)
	if err != nil {
		return "", err
	} else {
		return buffer.String(), nil
	}
}
