package storage

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/internal/s3mock"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func newStreamS3(t *testing.T, endpoint string) *S3 {
	v := viper.New()
	v.Set("bucket", "bk")
	v.Set("endpoint", endpoint)
	v.Set("path", "p")
	v.Set("part_size", 5) // MiB, the S3 minimum
	v.Set("access_key_id", "a")
	v.Set("secret_access_key", "b")
	base, err := newBase(config.ModelConfig{}, "", config.SubConfig{Type: "s3", Name: "s", Viper: v})
	assert.NoError(t, err)
	s := &S3{Base: base, Service: "s3"}
	assert.NoError(t, s.open())
	return s
}

type failAfter struct {
	r io.Reader
}

func (f *failAfter) Read(p []byte) (int, error) {
	n, err := f.r.Read(p)
	if err == io.EOF {
		return n, errors.New("dump failed")
	}
	return n, err
}

func Test_S3_uploadStream_multipart(t *testing.T) {
	srv := s3mock.New()
	defer srv.Close()
	s := newStreamS3(t, srv.URL)

	data := bytes.Repeat([]byte("0123456789"), 1200000) // 12 MB -> 3 parts
	assert.NoError(t, s.uploadStream(bytes.NewReader(data), "dump.sql"))

	got, ok := srv.Object("bk/p/dump.sql")
	assert.Equal(t, true, ok)
	assert.Equal(t, true, bytes.Equal(data, got))
	assert.Equal(t, 1, srv.Completed)
}

func Test_S3_uploadStream_small(t *testing.T) {
	srv := s3mock.New()
	defer srv.Close()
	s := newStreamS3(t, srv.URL)

	assert.NoError(t, s.uploadStream(bytes.NewReader([]byte("select 1;")), "a.sql"))
	got, _ := srv.Object("bk/p/a.sql")
	assert.Equal(t, "select 1;", string(got))
}

func Test_S3_uploadStream_readErrorAborts(t *testing.T) {
	srv := s3mock.New()
	defer srv.Close()
	s := newStreamS3(t, srv.URL)

	data := bytes.Repeat([]byte("x"), 12*1024*1024)
	err := s.uploadStream(&failAfter{r: bytes.NewReader(data)}, "bad.sql")
	assert.NotNil(t, err)
	_, ok := srv.Object("bk/p/bad.sql")
	assert.Equal(t, false, ok)
	assert.Equal(t, 0, srv.Completed)
	assert.Equal(t, 1, srv.Aborted)
}
