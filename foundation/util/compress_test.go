package util

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestZipFile(t *testing.T) string {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// 添加一个文件条目
	fw, err := w.Create("test.txt")
	assert.NoError(t, err)
	_, err = fw.Write([]byte("test content"))
	assert.NoError(t, err)

	// 添加一个目录条目
	_, err = w.Create("testdir/")
	assert.NoError(t, err)

	err = w.Close()
	assert.NoError(t, err)

	// 将缓冲区写入临时文件
	tmpfile, err := os.CreateTemp("", "testzip")
	assert.NoError(t, err)
	defer tmpfile.Close()

	_, err = tmpfile.Write(buf.Bytes())
	assert.NoError(t, err)

	return tmpfile.Name()
}

func TestUnZipFile_Success(t *testing.T) {
	zipFile := createTestZipFile(t)
	defer os.Remove(zipFile)

	destDir := t.TempDir()

	_, err := UnZipFile(zipFile, destDir)
	assert.NoError(t, err)

	// 验证文件和目录是否存在
	assert.FileExists(t, filepath.Join(destDir, "test.txt"))
	assert.DirExists(t, filepath.Join(destDir, "testdir"))
}

func TestUnZipFile_ZipFileOpenError(t *testing.T) {
	nonExistentZipFile := "non_existent.zip"
	destDir := t.TempDir()

	_, err := UnZipFile(nonExistentZipFile, destDir)
	assert.Error(t, err)
}

func TestZipDir_SuccessfulCompression(t *testing.T) {
	// 创建一个临时目录
	dir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// 在目录中创建一些文件
	files := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, file := range files {
		fullPath := filepath.Join(dir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// 创建一个临时文件用于输出 ZIP
	outputFile, err := os.CreateTemp("", "output.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outputFile.Name())

	// 调用 ZipDir 函数
	err = ZipDir(dir, outputFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// 验证 ZIP 文件的内容
	zipReader, err := zip.OpenReader(outputFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer zipReader.Close()

	// 检查 ZIP 文件中是否包含所有预期的文件
	for _, file := range files {
		found := false
		for _, zipFile := range zipReader.File {
			if zipFile.Name == file {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %s not found in ZIP", file)
		}
	}
}

func TestZipDir_FileCreationError(t *testing.T) {
	// 创建一个临时目录
	dir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// 创建一个不可写的输出文件
	outputFile := filepath.Join(dir, "output.zip")
	if err := os.WriteFile(outputFile, []byte{}, 0444); err != nil {
		t.Fatal(err)
	}

	// 调用 ZipDir 函数并期望它返回错误
	err = ZipDir(dir, outputFile)
	if err == nil {
		t.Fatal("Expected an error due to file creation failure")
	}
}

func TestZipDir_FileWalkError(t *testing.T) {
	// 创建一个临时目录
	dir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// 创建一个不可读的文件
	badFile := filepath.Join(dir, "badfile")
	if err := os.WriteFile(badFile, []byte{}, 0000); err != nil {
		t.Fatal(err)
	}

	// 创建一个临时文件用于输出 ZIP
	outputFile, err := os.CreateTemp("", "output.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outputFile.Name())

	// 调用 ZipDir 函数并期望它返回错误
	err = ZipDir(dir, outputFile.Name())
	if err == nil {
		t.Fatal("Expected an error due to file walk failure")
	}
}

func TestZipDir_FileOpenError(t *testing.T) {
	// 创建一个临时目录
	dir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// 创建一个不可读的文件
	badFile := filepath.Join(dir, "badfile")
	if err := os.WriteFile(badFile, []byte{}, 0000); err != nil {
		t.Fatal(err)
	}

	// 创建一个临时文件用于输出 ZIP
	outputFile, err := os.CreateTemp("", "output.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outputFile.Name())

	// 调用 ZipDir 函数并期望它返回错误
	err = ZipDir(dir, outputFile.Name())
	if err == nil {
		t.Fatal("Expected an error due to file open failure")
	}
}
