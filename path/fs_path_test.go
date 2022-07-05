package connector

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func Test_FSPath(t *testing.T) {
	testFolder := "/tmp/fs-shadow"
	_ = os.Mkdir(testFolder, os.ModePerm)

	emptyFolder := filepath.Join(testFolder, "test")
	_ = os.Mkdir(emptyFolder, os.ModePerm)

	emptyFile := filepath.Join(testFolder, "test.txt")
	_, _ = os.Create(emptyFile)

	filePath := NewFSPath(emptyFile)
	folderPath := NewFSPath(emptyFolder)

	assert.Equal(t, folderPath.IsDir(), true, "path is dir.")
	assert.Equal(t, folderPath.IsVirtual(), false, "path is not virtual.")
	assert.Equal(t, folderPath.Exists(), true, "path is exists.")
	assert.Equal(t, folderPath.Name(), "test", "invalid path name.")
	assert.Equal(t, folderPath.String(), emptyFolder, "invalid string.")
	assert.Equal(t, folderPath.ParentPath().Name(), "fs-shadow", "invalid parent name.")
	tmp := NewFSPath("/tmp")
	assert.Equal(t, folderPath.ExcludePath(tmp).String(), "fs-shadow/test", "invalid folder name.")

	assert.Equal(t, filePath.IsDir(), false, "file path is file.")
	assert.Equal(t, filePath.IsVirtual(), false, "file path is not virtual.")
	assert.Equal(t, filePath.Exists(), true, "file path is exists.")
	assert.Equal(t, filePath.Name(), "test.txt", "invalid file path name.")
	assert.Equal(t, filePath.String(), emptyFile, "invalid string.")
	assert.Equal(t, filePath.ParentPath().Name(), "fs-shadow", "invalid parent file name.")
	tmp = NewFSPath("/tmp")
	assert.Equal(t, folderPath.ExcludePath(tmp).String(), "fs-shadow/test", "invalid file name.")

	_ = os.RemoveAll(testFolder)
}
