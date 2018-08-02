package net

import "bytes"

// Simple hack for replacing a ReadCloser
// with an in-memory buffer

type MemReadCloser struct{
	b *bytes.Reader
}

func NewMemReadCloser(buf []byte) MemReadCloser {
	return MemReadCloser{ bytes.NewReader(buf) }
}

func (m MemReadCloser) Read(b []byte) (int, error) {
	return m.b.Read(b)
}

func (m MemReadCloser) Close() error {
	return nil
}
