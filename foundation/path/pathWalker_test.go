package path

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	dir := filepath.Join(os.TempDir(), "path_test")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// 创建测试文件
	files := []string{
		"test1.txt",
		"test2.txt",
		"subdir/test3.txt",
		"subdir/test4.log",
		"test5.log",
	}

	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Create(path); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func cleanupTestDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("cleanup error: %v", err)
	}
}

func TestListPath(t *testing.T) {
	testDir := setupTestDir(t)
	defer cleanupTestDir(t, testDir)

	tests := []struct {
		name         string
		path         string
		filter       string
		recursive    bool
		expectedLen  int
		expectError  bool
	}{
		{
			name:        "non-recursive txt files",
			path:        testDir,
			filter:      "*.txt",
			recursive:   false,
			expectedLen: 2,
		},
		{
			name:        "recursive txt files",
			path:        testDir,
			filter:      "*.txt",
			recursive:   true,
			expectedLen: 3,
		},
		{
			name:        "non-recursive log files",
			path:        testDir,
			filter:      "*.log",
			recursive:   false,
			expectedLen: 1,
		},
		{
			name:        "recursive all files",
			path:        testDir,
			filter:      "*",
			recursive:   true,
			expectedLen: 5,
		},
		{
			name:        "invalid path",
			path:        filepath.Join(testDir, "nonexistent"),
			filter:      "*",
			recursive:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := ListPath(tt.path, tt.filter, tt.recursive)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(files) != tt.expectedLen {
				t.Errorf("expected %d files, got %d", tt.expectedLen, len(files))
			}
		})
	}
}