package net

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type interruptedBody struct{ io.Reader }

func (r interruptedBody) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return n, err
}

func TestBodyFilePublishesCompleteBytesAndPreservesPreviousOnFailure(t *testing.T) {
	for _, size := range []int{0, 10 << 20, (10 << 20) + 1, 11 << 20} {
		dir := t.TempDir()
		body := bytes.Repeat([]byte("x"), size)
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		if err := HTTPBodyToFile(req, dir, "file"); err != nil {
			t.Fatal(err)
		}
		got, _ := os.ReadFile(filepath.Join(dir, "file"))
		if !bytes.Equal(got, body) {
			t.Fatalf("size %d was truncated to %d", size, len(got))
		}
	}
	for _, kind := range []string{"interrupted", "short", "limit", "cancelled", "publish-failed"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			destination := filepath.Join(dir, "file")
			if err := os.WriteFile(destination, []byte("old"), 0600); err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("new")))
			switch kind {
			case "interrupted":
				req.Body = io.NopCloser(interruptedBody{bytes.NewReader([]byte("new"))})
			case "short":
				req.ContentLength = 4
			case "limit":
				req.Body = http.MaxBytesReader(httptest.NewRecorder(), req.Body, 2)
			case "cancelled":
				ctx, cancel := context.WithCancel(req.Context())
				cancel()
				req = req.WithContext(ctx)
			case "publish-failed":
				_ = os.Remove(destination)
				_ = os.Mkdir(destination, 0700)
			}
			if err := HTTPBodyToFile(req, dir, "file"); err == nil {
				t.Fatal("failure returned success")
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 1 {
				t.Fatal("temporary file leaked", entries)
			}
			if kind != "publish-failed" {
				got, _ := os.ReadFile(destination)
				if string(got) != "old" {
					t.Fatal("old contents destroyed")
				}
			}
		})
	}
}

func TestMultipartFileHasNoImplicitBinaryLimit(t *testing.T) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, _ := w.CreateFormFile("file", "file")
	want := bytes.Repeat([]byte("x"), 11<<20)
	_, _ = part.Write(want)
	_ = w.Close()
	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if err := req.ParseMultipartForm(1024); err != nil {
		t.Fatal(err)
	}
	defer req.MultipartForm.RemoveAll()
	dir := t.TempDir()
	if _, err := MultipartFormFile(req, "file", dir, ""); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, "file"))
	if !bytes.Equal(got, want) {
		t.Fatal("multipart was truncated")
	}
}

func TestFileRejectsTraversalAndFramingMismatch(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{".", "..", "../file", "bad\x00name"} {
		if err := WriteFileAtomic(context.Background(), bytes.NewReader(nil), dir, name, 0); err == nil {
			t.Fatal("invalid name accepted", name)
		}
	}
	err := WriteFileAtomic(context.Background(), bytes.NewReader([]byte("x")), dir, "file", 2)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal(err)
	}
}

type failingFileWriter struct {
	writeErr, syncErr, closeErr error
	closed                      bool
}

func (s *failingFileWriter) Write(p []byte) (int, error) {
	if s.writeErr != nil {
		return 0, s.writeErr
	}
	return len(p), nil
}
func (s *failingFileWriter) Sync() error  { return s.syncErr }
func (s *failingFileWriter) Close() error { s.closed = true; return s.closeErr }
func TestFileWriteReportsStorageFailuresAndAlwaysCloses(t *testing.T) {
	failure := errors.New("storage failure")
	for _, writer := range []*failingFileWriter{{writeErr: failure}, {syncErr: failure}, {closeErr: failure}} {
		if err := copyFileContent(context.Background(), writer, bytes.NewBufferString("complete"), 8); !errors.Is(err, failure) {
			t.Fatal("storage failure hidden", err)
		}
		if !writer.closed {
			t.Fatal("destination not closed")
		}
	}
}
