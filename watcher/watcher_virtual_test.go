package watcher

import (
	"github.com/ayhanozemre/fs-shadow/filenode"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func Test_VirtualWatcherUseCase(t *testing.T) {
	root := "fs-shadow"
	watcher, err := NewVirtualPathWatcher(root)
	assert.Equal(t, nil, err, "watcher creation error")

	watcher.Close()
}

func Test_VirtualWatcherFunctionality(t *testing.T) {
	var err error
	parentPath := "/tmp"
	testRoot := filepath.Join(parentPath, "fs-shadow")

	path := connector.NewVirtualPath(testRoot, true)

	root := filenode.FileNode{
		Name: path.Name(),
		Meta: filenode.MetaData{
			IsDir: true,
		},
	}

	tw := VirtualTree{
		FileTree:   &root,
		ParentPath: path.ParentPath(),
		Path:       path,
	}

	err = tw.Create(path)
	assert.Equal(t, nil, err, "root node creation error")

	// Create folder
	newFolder := connector.NewVirtualPath(filepath.Join(testRoot, "folder"), true)
	err = tw.Create(newFolder)
	assert.Equal(t, nil, err, "folder node creation error")
	assert.Equal(t, newFolder.Name(), tw.FileTree.Subs[0].Name, "create:invalid folder name")

	// Create file
	newFile := connector.NewVirtualPath(filepath.Join(testRoot, "file.txt"), false)
	err = tw.Create(newFile)
	assert.Equal(t, nil, err, "file node creation error")
	assert.Equal(t, newFile.Name(), tw.FileTree.Subs[1].Name, "create:invalid file name")

	// Rename
	renameFilePath := connector.NewVirtualPath(filepath.Join(testRoot, "file-rename.txt"), false)
	err = tw.Rename(newFile, renameFilePath)
	assert.Equal(t, nil, err, "file node rename error")
	assert.Equal(t, renameFilePath.Name(), tw.FileTree.Subs[1].Name, "rename:filename is not changed")

	// Write
	err = tw.Write(renameFilePath)
	assert.Equal(t, nil, err, "file node write error")

	// Remove
	err = tw.Remove(renameFilePath)
	assert.Equal(t, nil, err, "file node remove error")
	assert.Equal(t, 1, len(root.Subs), "file node not removed")

}
