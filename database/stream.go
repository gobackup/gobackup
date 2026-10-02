package database

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/gobackup/gobackup/config"
)

// Stream is a database dump read directly from the dump process stdout.
//
// Read returns the process error instead of io.EOF when the process exits
// with a non-zero status, so the consumer never mistakes a failed dump for a
// complete one.
type Stream struct {
	// Name is the dump file name, e.g. "mydb.sql".
	Name string
	// Ext is the extension to use for the uploaded file, e.g. ".sql".
	Ext string

	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr *bytes.Buffer
	done   bool
}

// Streamer is implemented by databases that can dump to a stream.
type Streamer interface {
	stream() (*Stream, error)
}

func startStream(command, name, ext string) (*Stream, error) {
	parts := strings.Fields(command)
	fullCommand, err := exec.LookPath(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%s cannot be found", parts[0])
	}

	cmd := exec.Command(fullCommand, parts[1:]...)
	cmd.Env = os.Environ()
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Stream{Name: name, Ext: ext, cmd: cmd, stdout: stdout, stderr: stderr}, nil
}

func (s *Stream) Read(p []byte) (int, error) {
	n, err := s.stdout.Read(p)
	if err == io.EOF && !s.done {
		s.done = true
		if werr := s.cmd.Wait(); werr != nil {
			msg := strings.TrimSpace(s.stderr.String())
			// dump tools log progress to stderr, keep only the tail
			if len(msg) > 1000 {
				msg = "..." + msg[len(msg)-1000:]
			}
			return n, fmt.Errorf("dump error: %s", msg)
		}
	}
	return n, err
}

// Close kills the dump process if it is still running.
func (s *Stream) Close() error {
	if s.done {
		return nil
	}
	s.done = true
	_ = s.cmd.Process.Kill()
	_ = s.cmd.Wait()
	return nil
}

// RunStream starts the dump of the model's only database as a stream.
func RunStream(model config.ModelConfig) (*Stream, error) {
	if len(model.Databases) != 1 {
		return nil, errors.New("stream mode requires exactly one database")
	}

	for _, dbCfg := range model.Databases {
		base := newBase(model, dbCfg)
		var db Database
		switch dbCfg.Type {
		case "mysql":
			db = &MySQL{Base: base}
		case "mongodb":
			db = &MongoDB{Base: base}
		case "postgresql":
			db = &PostgreSQL{Base: base}
		default:
			return nil, fmt.Errorf("database type %q does not support stream", dbCfg.Type)
		}

		if err := runHook("dump before_script", dbCfg.Viper.GetString("before_script")); err != nil {
			return nil, err
		}
		if err := db.init(); err != nil {
			return nil, err
		}
		return db.(Streamer).stream()
	}
	return nil, nil
}
