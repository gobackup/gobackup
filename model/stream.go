package model

import (
	"compress/gzip"
	"io"
)

// countReader counts bytes read from r.
type countReader struct {
	r io.Reader
	n int64
}

func (c *countReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// gzipStream compresses r on the fly. Errors from r are propagated to the reader.
// Closing the returned reader unblocks and stops the compressing goroutine.
func gzipStream(r io.Reader) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		zw := gzip.NewWriter(pw)
		_, err := io.Copy(zw, r)
		if cerr := zw.Close(); err == nil {
			err = cerr
		}
		pw.CloseWithError(err)
	}()
	return pr
}
