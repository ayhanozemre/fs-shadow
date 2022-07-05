package filenode

import (
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func Test_WalkOnFsPath(t *testing.T) {
	testFolder := "/tmp/fs-shadow"
	rootPath := connector.NewFSPath(testFolder)

	_ = os.Mkdir(testFolder, os.ModePerm)
	folder := filepath.Join(testFolder, "test")
	_ = os.Mkdir(folder, os.ModePerm)
	_, _ = os.Create(filepath.Join(folder, "sub.txt"))

	root := FileNode{
		Name: rootPath.Name(),
		Meta: MetaData{
			IsDir: true,
		},
	}
	var directoryCount int
	var wg sync.WaitGroup
	eventCh := make(chan connector.Path)
	go func() {
		for {
			select {
			case p := <-eventCh:
				if p == nil {
					return
				} else {
					directoryCount += 1
				}
			}
		}
	}()
	WalkOnFsPath(&root, rootPath, &wg, eventCh)
	wg.Wait()
	eventCh <- nil
	close(eventCh)
	assert.Equal(t, 2, directoryCount, "directory count is not equal to expected count")

	assert.Equal(t, root.Subs[0].Name, "test", "mismatch sub folder name error")
	assert.Equal(t, root.Subs[0].Subs[0].Name, "sub.txt", "mismatch sub file name error")
	_ = os.RemoveAll(testFolder)
}

func Test_FileNode(t *testing.T) {
	var err error
	parentPath := "/tmp"
	testFolder := filepath.Join(parentPath, "fs-shadow")
	_ = os.Mkdir(testFolder, os.ModePerm)
	rootPath := connector.NewFSPath(testFolder)

	folder := filepath.Join(testFolder, "test")
	_ = os.Mkdir(folder, os.ModePerm)
	_, _ = os.Create(filepath.Join(folder, "sub.txt"))
	folderPath := connector.NewFSPath(folder)
	eventFolderPath := folderPath.ExcludePath(connector.NewFSPath(parentPath))

	emptyFile := filepath.Join(testFolder, "test.txt")
	_, _ = os.Create(emptyFile)
	filePath := connector.NewFSPath(emptyFile)
	eventFilePath := filePath.ExcludePath(connector.NewFSPath(parentPath))

	renameFilePath := connector.NewFSPath(filepath.Join(testFolder, "test-2.txt"))
	renameEventFilePath := renameFilePath.ExcludePath(connector.NewFSPath(parentPath))

	root := FileNode{
		Name: rootPath.Name(),
		Meta: MetaData{
			IsDir: true,
		},
	}

	// creation
	eventCh := make(chan connector.Path)
	go func() {
		for {
			select {
			case p := <-eventCh:
				if p == nil {
					return
				}
			}
		}
	}()

	err = root.Create(eventFolderPath, folderPath, eventCh)
	assert.Equal(t, nil, err, "folder creation error")
	folderNode := root.Subs[0]
	assert.Equal(t, true, folderNode.Meta.IsDir, "folder node is not dir")
	assert.Equal(t, "eae6903bedc6d6aef6eb50f23865dd544469f83e9662171081881a23f1fc79b3", folderNode.Meta.Sum, "invalid folder sum")

	err = root.Create(eventFilePath, filePath, eventCh)
	assert.Equal(t, nil, err, "file creation error")
	fileNode := root.Subs[1]
	assert.Equal(t, false, fileNode.Meta.IsDir, "file node is not file")
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", fileNode.Meta.Sum, "invalid file sum")

	eventCh <- nil
	close(eventCh)

	// Search
	node := root.Search(eventFilePath.String())
	assert.NotEqual(t, nil, node, "search file error")
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", node.Meta.Sum, "wrong node found")

	// Update
	oldSum := fileNode.Meta.Sum
	f, _ := os.OpenFile(emptyFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	_, err = f.WriteString("test")
	_ = f.Close()
	err = root.Update(eventFilePath, filePath)
	assert.Equal(t, nil, err, "update error")
	assert.NotEqual(t, oldSum, fileNode.Meta.Sum, "updated file sums not equal")

	// Rename
	oldName := fileNode.Name
	err = root.Rename(eventFilePath, renameEventFilePath)
	assert.Equal(t, nil, err, "rename error")
	assert.NotEqual(t, oldName, fileNode.Name, "rename process error")
	_ = os.Rename(filePath.String(), renameFilePath.String())

	// Remove
	err, _ = root.Remove(eventFolderPath)
	assert.Equal(t, nil, err, "remove error")
	deletedNode := root.Search(eventFolderPath.String())
	assert.Nil(t, deletedNode, "remove process error")

	// SumUpdate
	oldSum = node.Meta.Sum

	f, _ = os.OpenFile(renameFilePath.String(), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	_, err = f.WriteString("rename2")
	_ = f.Close()

	node = root.Search(renameEventFilePath.String())
	err = node.SumUpdate(renameFilePath)
	assert.Equal(t, nil, err, "sum update error")
	assert.NotEqual(t, oldSum, node.Meta.Sum, "sums not equal")

	_ = os.RemoveAll(testFolder)
}
