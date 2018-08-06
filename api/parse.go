package api

import (
	"strconv"
	"strings"
	"errors"
)

var durationErr = errors.New("unknown duration code")

// "137,802 views" => 137802
func ExtractNumber(s string) (uint64, error) {
	// Extract numbers from view string
	var clean []byte
	for _, char := range []byte(s) {
		if char >= 0x30 && char <= 0x39 {
			clean = append(clean, char)
		}
	}

	// Convert to uint
	return strconv.ParseUint(string(clean), 10, 64)
}

// "PT6M57S" => 6 min 57 s
func ParseDuration(d string) (uint64, error) {
	if d[0:2] != "PT" { return 0, durationErr }
	mIndex := strings.IndexByte(d, 'M')
	if mIndex == -1 { return 0, durationErr }

	minutes, err := strconv.ParseUint(d[2:mIndex], 10, 32)
	if err != nil { return 0, err }
	seconds, err := strconv.ParseUint(d[mIndex+1:len(d)-1], 10, 32)
	if err != nil { return 0, err }

	dur := minutes * 60 + seconds
	return dur, nil
}
