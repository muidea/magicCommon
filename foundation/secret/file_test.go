package secret

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadRejectsUnsafeFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "value")
	if err := os.WriteFile(path, []byte("value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path); err == nil {
		t.Fatal("Read accepted group/world-readable secret")
	}
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	value, err := Read(path)
	if err != nil || string(value) != "value" {
		t.Fatalf("Read() = %q, %v", value, err)
	}
}

func TestWriteOnceDoesNotReplaceCredential(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential.json")
	if err := WriteOnce(path, []byte(`{"clientId":"client-1"}`)); err != nil {
		t.Fatal(err)
	}
	if err := WriteOnce(path, []byte(`{"clientId":"client-2"}`)); err == nil {
		t.Fatal("WriteOnce replaced an existing credential")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = %o, want 600", info.Mode().Perm())
	}
}
