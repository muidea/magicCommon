package net

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteFileAtomic publishes only closed, complete bytes. Limits belong to the
// caller; expectedSize verifies framing, including readers that stop early.
func WriteFileAtomic(ctx context.Context, reader io.Reader, directory, name string, expectedSize int64) (err error) {
	if ctx == nil || reader == nil || !isValidFileName(name) || !isValidDirectory(directory) {
		return errors.New("invalid file destination or reader")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	f, err := os.CreateTemp(directory, ".upload-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if err = copyFileContent(ctx, f, reader, expectedSize); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(directory, name))
}

// copyFileContent closes the destination on every outcome, including partial
// writes, storage synchronization failures and interrupted readers.
func copyFileContent(ctx context.Context, destination interface {
	io.WriteCloser
	Sync() error
}, reader io.Reader, expectedSize int64) error {
	n, copyErr := io.Copy(destination, reader)
	syncErr := destination.Sync()
	closeErr := destination.Close()
	if err := errors.Join(copyErr, syncErr, closeErr, ctx.Err()); err != nil {
		return err
	}
	if expectedSize >= 0 && n != expectedSize {
		return fmt.Errorf("incomplete file: copied %d bytes, expected %d: %w", n, expectedSize, io.ErrUnexpectedEOF)
	}
	return nil
}
