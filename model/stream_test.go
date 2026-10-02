package model

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/longbridgeapp/assert"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func Test_gzipStream(t *testing.T) {
	cr := &countReader{r: strings.NewReader("hello stream")}
	zr, err := gzip.NewReader(gzipStream(cr))
	assert.NoError(t, err)
	out, err := io.ReadAll(zr)
	assert.NoError(t, err)
	assert.Equal(t, "hello stream", string(out))
	assert.Equal(t, int64(12), cr.n)
}

func Test_gzipStream_propagatesError(t *testing.T) {
	_, err := io.Copy(&bytes.Buffer{}, gzipStream(errReader{}))
	assert.NotNil(t, err)
}
