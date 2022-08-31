package watcher

import (
	"context"
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_WatcherUseCase(t *testing.T) {
	testRoot := "/tmp/fs-shadow"
	_ = os.Mkdir(testRoot, os.ModePerm)
	tw, _, err := NewPathWatcher(testRoot)
	assert.Equal(t, nil, err, "path watcher creation error")

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
	time.Sleep(3 * time.Second)
	assert.Equal(t, 0, len(tw.FileTree.Subs), "remove:invalid subs length")

	tw.Stop()
	_ = os.RemoveAll(testRoot)
	_ = os.Remove(moveDirectory)

}

func Test_WatcherFunctionality(t *testing.T) {
	var err error
	var watcher *fsnotify.Watcher
	parentPath := "/tmp"
	testRoot := filepath.Join(parentPath, "fs-shadow")
	_ = os.Mkdir(testRoot, os.ModePerm)

	path := connector.NewFSPath(testRoot)

	watcher, err = fsnotify.NewWatcher()
	assert.Equal(t, nil, err, "watcher creation error")

	root := filenode.FileNode{
		Name: path.Name(),
		Meta: filenode.MetaData{
			IsDir: true,
		},
	}

	tw := TreeWatcher{
		FileTree:     &root,
		ParentPath:   path.ParentPath(),
		Path:         path,
		Watcher:      watcher,
		EventManager: event.NewEventHandler(),
	}
	tw.IgniterReloadCtx, tw.IgniterReloadFunc = context.WithCancel(context.Background())

	_, err = tw.Create(path, nil)
	assert.Equal(t, nil, err, "root node creation error")

	// Create folder
	newFolder := connector.NewFSPath(filepath.Join(testRoot, "folder"))
	_ = os.Mkdir(newFolder.String(), os.ModePerm)
	_, err = tw.Create(newFolder, nil)
	assert.Equal(t, nil, err, "folder node creation error")
	assert.Equal(t, newFolder.Name(), tw.FileTree.Subs[0].Name, "create:invalid folder name")

	// Create file
	newFile := connector.NewFSPath(filepath.Join(testRoot, "file.txt"))
	_, _ = os.Create(newFile.String())
	_, err = tw.Create(newFile, nil)
	assert.Equal(t, nil, err, "file node creation error")
	assert.Equal(t, newFile.Name(), tw.FileTree.Subs[1].Name, "create:invalid file name")

	// Rename
	renameFilePath := connector.NewFSPath(filepath.Join(testRoot, "file-rename.txt"))
	_ = os.Rename(newFile.String(), renameFilePath.String())
	_, err = tw.Rename(newFile, renameFilePath)
	assert.Equal(t, nil, err, "file node rename error")
	assert.Equal(t, renameFilePath.Name(), tw.FileTree.Subs[1].Name, "rename:filename is not changed")

	// Write
	renameEventFilePath := renameFilePath.ExcludePath(connector.NewFSPath(parentPath))
	node := tw.FileTree.Search(renameEventFilePath.String())
	assert.NotEqual(t, nil, node, "renamed file not found")
	oldSum := node.Meta.Sum

	f, _ := os.OpenFile(renameFilePath.String(), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	_, err = f.WriteString("test")
	_ = f.Close()

	_, err = tw.Write(renameFilePath)
	assert.Equal(t, nil, err, "file node write error")
	node = tw.FileTree.Search(renameEventFilePath.String())
	assert.NotEqual(t, oldSum, node.Meta.Sum, "updated file sums not equal")

	// Remove
	_, err = tw.Remove(renameFilePath)
	assert.Equal(t, nil, err, "file node remove error")
	assert.Equal(t, 1, len(root.Subs), "file node not removed")

	var e event.Event
	// Handler Create
	handlerTestFile := connector.NewFSPath(filepath.Join(testRoot, "new-file.txt"))
	_, _ = os.Create(handlerTestFile.String())
	e = event.Event{FromPath: handlerTestFile, Type: event.Create}
	_, err = tw.Handler(e, nil)
	assert.Equal(t, nil, err, "handler creation error")
	assert.Equal(t, handlerTestFile.Name(), tw.FileTree.Subs[1].Name, "handler: filename mismatch error")

	// Handler Write
	oldSum = tw.FileTree.Subs[1].Meta.Sum
	f, _ = os.OpenFile(handlerTestFile.String(), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	_, err = f.WriteString("handler-write")
	_ = f.Close()

	e = event.Event{FromPath: handlerTestFile, Type: event.Write}
	_, err = tw.Handler(e, nil)
	assert.Equal(t, nil, err, "handler write error")
	assert.NotEqual(t, oldSum, tw.FileTree.Subs[1].Meta.Sum, "handler: file sum mismatch")

	// Handler Rename
	handlerRenameTestFile := connector.NewFSPath(filepath.Join(testRoot, "new-file-rename.txt"))
	_ = os.Rename(handlerTestFile.String(), handlerRenameTestFile.String())
	e = event.Event{FromPath: handlerTestFile, ToPath: handlerRenameTestFile, Type: event.Rename}
	_, err = tw.Handler(e, nil)
	assert.Equal(t, nil, err, "handler rename error")
	assert.Equal(t, handlerRenameTestFile.Name(), tw.FileTree.Subs[1].Name, "handler: filename mismatch error")

	// Handler Remove
	handlerRenameTestFile = connector.NewFSPath(filepath.Join(testRoot, "new-file-rename.txt"))
	_ = os.Remove(handlerRenameTestFile.String())
	e = event.Event{FromPath: handlerRenameTestFile, Type: event.Remove}
	_, err = tw.Handler(e, nil)
	assert.Equal(t, nil, err, "handler remove error")
	assert.Equal(t, len(tw.FileTree.Subs), 1, "handler: root subs length error")

	_ = os.RemoveAll(testRoot)
}
