// Package s3mock is a minimal in-memory, path-style S3 server for tests.
// It supports PUT/GET/DELETE object and multipart upload.
package s3mock

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	*httptest.Server

	mu        sync.Mutex
	objects   map[string][]byte
	uploads   map[string]map[int][]byte
	nextID    int
	Completed int
	Aborted   int
}

func New() *Server {
	s := &Server{objects: map[string][]byte{}, uploads: map[string]map[int][]byte{}}
	s.Server = httptest.NewServer(http.HandlerFunc(s.handle))
	return s
}

// Object returns the object stored under "bucket/key".
func (s *Server) Object(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.objects[key]
	return b, ok
}

// Keys returns all stored object keys, sorted.
func (s *Server) Keys() (keys []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.TrimPrefix(r.URL.Path, "/")
	q := r.URL.Query()
	uploadID := q.Get("uploadId")
	_, isUploads := q["uploads"]

	switch {
	case r.Method == http.MethodPost && isUploads:
		s.nextID++
		id := strconv.Itoa(s.nextID)
		s.uploads[id] = map[int][]byte{}
		bucket, k, _ := strings.Cut(key, "/")
		fmt.Fprintf(w, `<?xml version="1.0"?><InitiateMultipartUploadResult><Bucket>%s</Bucket><Key>%s</Key><UploadId>%s</UploadId></InitiateMultipartUploadResult>`, bucket, k, id)
	case r.Method == http.MethodPut && uploadID != "":
		n, _ := strconv.Atoi(q.Get("partNumber"))
		b, _ := io.ReadAll(r.Body)
		s.uploads[uploadID][n] = b
		w.Header().Set("ETag", fmt.Sprintf(`"part-%d"`, n))
	case r.Method == http.MethodPost && uploadID != "":
		parts := s.uploads[uploadID]
		var nums []int
		for n := range parts {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		var all []byte
		for _, n := range nums {
			all = append(all, parts[n]...)
		}
		s.objects[key] = all
		delete(s.uploads, uploadID)
		s.Completed++
		bucket, k, _ := strings.Cut(key, "/")
		fmt.Fprintf(w, `<?xml version="1.0"?><CompleteMultipartUploadResult><Location>%s</Location><Bucket>%s</Bucket><Key>%s</Key><ETag>"done"</ETag></CompleteMultipartUploadResult>`, r.URL.Path, bucket, k)
	case r.Method == http.MethodDelete && uploadID != "":
		delete(s.uploads, uploadID)
		s.Aborted++
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodPut:
		b, _ := io.ReadAll(r.Body)
		s.objects[key] = b
		w.Header().Set("ETag", `"single"`)
	case r.Method == http.MethodGet:
		b, ok := s.objects[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `<Error><Code>NoSuchKey</Code></Error>`)
			return
		}
		w.Write(b)
	case r.Method == http.MethodDelete:
		delete(s.objects, key)
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotImplemented)
	}
}
