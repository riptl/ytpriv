package classic

import (
	"time"
	"errors"
	"strings"
	"strconv"
)

// "PT6M57S" => 6 min 57 s
func parseDuration(d string) (time.Duration, error) {
	var err error
	goto start

error:
	return 0, errors.New("unknown duration code")

start:
	if d[0:2] != "PT" { goto error }
	mIndex := strings.IndexByte(d, 'M')
	if mIndex == -1 { goto error }

	minutes, err := strconv.ParseUint(d[2:mIndex], 10, 32)
	if err != nil { return 0, err }
	seconds, err := strconv.ParseUint(d[mIndex:len(d)-1], 10, 32)
	if err != nil { return 0, err }

	dur := time.Duration(minutes) * time.Minute + time.Duration(seconds) * time.Second
	return dur, nil
}

// "137,802 views" => 137802
func extractNumber(s string) (uint64, error) {
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
