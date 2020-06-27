package api

import (
	"bytes"
	"encoding/base64"
	"strconv"
)

func GenChannelPageToken(channelId string, page uint64) string {
	// Generate the inner token
	token := genInnerToken(page)

	// Build the inner object
	var inner bytes.Buffer

	// channelId
	inner.WriteByte(0x12)                       // type
	writeVarint(&inner, uint64(len(channelId))) // len
	inner.WriteString(channelId)                // data

	// token
	inner.WriteByte(0x1a)                   // type
	writeVarint(&inner, uint64(len(token))) // len
	inner.WriteString(token)                // data

	innerBytes := inner.Bytes()

	var root bytes.Buffer

	// innerBytes
	root.Write([]byte{0xe2, 0xa9, 0x85, 0xb2, 0x02}) // probably types
	writeVarint(&root, uint64(len(innerBytes)))
	root.Write(innerBytes)

	rootBytes := root.Bytes()

	return base64.URLEncoding.EncodeToString(rootBytes)
}

func genInnerToken(page uint64) string {
	var buf bytes.Buffer

	pageStr := strconv.FormatUint(page, 10)

	// Probably protobuf
	buf.Write([]byte{0x12, 0x06})
	buf.WriteString("videos")
	buf.Write([]byte{
		0x20, 0x00, 0x30, 0x01, 0x38, 0x01, 0x60, 0x01,
		0x6a, 0x00, 0x7a,
	})
	// Write size-prefixed page string
	writeVarint(&buf, uint64(len(pageStr)))
	buf.WriteString(pageStr)
	buf.Write([]byte{0xb8, 0x01, 0x00})

	return base64.URLEncoding.EncodeToString(buf.Bytes())
}

func writeVarint(buf *bytes.Buffer, n uint64) {
	var enc [10]byte
	i := uint(0)
	for {
		enc[i] = uint8(n & 0x7F)
		n >>= 7
		if n != 0 {
			enc[i] |= 0x80
			i++
		} else {
			i++
			break
		}
	}
	buf.Write(enc[:i])
}
