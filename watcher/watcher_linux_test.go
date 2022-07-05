package watcher

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_UseCase(t *testing.T) {
	testRoot := "/tmp/fs-shadow"
	_ = os.Mkdir(testRoot, os.ModePerm)

	tw, err := newLinuxPathWatcher(testRoot)
	assert.Equal(t, nil, err, "linux patch watcher creation error")

	// create folder
	folderName := "test1"
	folder := filepath.Join(testRoot, folderName)
	_ = os.Mkdir(folder, os.ModePerm)
	time.Sleep(3 * time.Second)
	assert.Equal(t, folderName, tw.FileTree.Subs[0].Name, "create:invalid folder name")

	// rename folder
	newFolderName := "test1-rename"
	renameFolder := filepath.Join(testRoot, newFolderName)
	_ = os.Rename(folder, renameFolder)
	time.Sleep(2 * time.Second)
	assert.Equal(t, newFolderName, tw.FileTree.Subs[0].Name, "rename:invalid folder name")

	// move to other directory
	moveDirectory := "/tmp/test1-rename"
	err = os.Rename(renameFolder, moveDirectory)
	time.Sleep(2 * time.Second)
	assert.Equal(t, 0, len(tw.FileTree.Subs), "remove:invalid subs length")

	tw.Close()
	_ = os.RemoveAll(testRoot)
	_ = os.Remove(moveDirectory)

}

func Test_Functionality(t *testing.T) {

}
