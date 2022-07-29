package watcher

import (
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func Test_VirtualWatcherUseCase(t *testing.T) {
	root := "fs-shadow"
	tw, _, err := NewVirtualPathWatcher(root, &filenode.ExtraPayload{UUID: uuid.NewString()})
	assert.Equal(t, nil, err, "watcher creation error")

	newPath := connector.NewVirtualPath(filepath.Join(root, "test-1"), true)
	// Create
	e := event.Event{FromPath: newPath, Type: event.Create}
	_, err = tw.Handler(e, &filenode.ExtraPayload{UUID: uuid.NewString()})
	assert.Equal(t, nil, err, "folder creation error")
	assert.Equal(t, newPath.Name(), tw.FileTree.Subs[0].Name, "create:invalid file name")

	// Rename
	renameNewPath := connector.NewVirtualPath(filepath.Join(root, "test-2"), true)
	e = event.Event{FromPath: newPath, ToPath: renameNewPath, Type: event.Rename}
	_, err = tw.Handler(e, nil)
	assert.Equal(t, nil, err, "folder rename error")
	assert.Equal(t, renameNewPath.Name(), tw.FileTree.Subs[0].Name, "rename:invalid file name")

	// Remove
	e = event.Event{FromPath: renameNewPath, Type: event.Remove}
	_, err = tw.Handler(e, nil)
	assert.Equal(t, nil, err, "folder remove error")
	assert.Equal(t, 0, len(tw.FileTree.Subs), "file node not removed")

}

func Test_VirtualWatcherFunctionality(t *testing.T) {
	var err error
	parentPath := "/tmp"
	testRoot := filepath.Join(parentPath, "fs-shadow")

	path := connector.NewVirtualPath(testRoot, true)

	root := filenode.FileNode{
		Name: path.Name(),
		UUID: uuid.NewString(),
		Meta: filenode.MetaData{
			IsDir: true,
		},
	}

	tw := VirtualTree{
		FileTree:   &root,
		ParentPath: path.ParentPath(),
		Path:       path,
	}

	_, err = tw.Create(path, &filenode.ExtraPayload{UUID: uuid.NewString()})
	assert.Equal(t, nil, err, "root node creation error")

	// Create folder
	newFolder := connector.NewVirtualPath(filepath.Join(testRoot, "folder"), true)
	_, err = tw.Create(newFolder, &filenode.ExtraPayload{UUID: uuid.NewString()})
	assert.Equal(t, nil, err, "folder node creation error")
	assert.Equal(t, newFolder.Name(), tw.FileTree.Subs[0].Name, "create:invalid folder name")

	// Create file
	newFile := connector.NewVirtualPath(filepath.Join(testRoot, "file.txt"), false)
	_, err = tw.Create(newFile, &filenode.ExtraPayload{UUID: uuid.NewString()})
	assert.Equal(t, nil, err, "file node creation error")
	assert.Equal(t, newFile.Name(), tw.FileTree.Subs[1].Name, "create:invalid file name")

	// Rename
	renameFilePath := connector.NewVirtualPath(filepath.Join(testRoot, "file-rename.txt"), false)
	_, err = tw.Rename(newFile, renameFilePath)
	assert.Equal(t, nil, err, "file node rename error")
	assert.Equal(t, renameFilePath.Name(), tw.FileTree.Subs[1].Name, "rename:filename is not changed")

	// Write
	_, err = tw.Write(renameFilePath)
	assert.Equal(t, nil, err, "file node write error")

	// Move
	_, err = tw.Move(renameFilePath, newFolder)
	assert.Equal(t, nil, err, "file node move error")
	assert.Equal(t, 1, len(root.Subs), "file node move error")

	// Remove
	_, err = tw.Remove(renameFilePath)
	assert.Equal(t, nil, err, "file node remove error")
	assert.Equal(t, 1, len(root.Subs), "file node not removed")
}

func Test_Restore(t *testing.T) {
	root := "fs-shadow"
	tw, _, err := NewVirtualPathWatcher(root, &filenode.ExtraPayload{UUID: uuid.NewString()})
	assert.Equal(t, nil, err, "watcher creation error")
	tw.Restore(&filenode.FileNode{Name: "new-tree"})
	assert.Equal(t, "new-tree", tw.FileTree.Name, "tree updated error")

}

func TestVirtualTree_SearchByPath(t *testing.T) {
	root := "fs-shadow"
	tw, _, _ := NewVirtualPathWatcher(root, &filenode.ExtraPayload{UUID: uuid.NewString()})
	newPath := connector.NewVirtualPath(filepath.Join(root, "test-1"), true)
	e := event.Event{FromPath: newPath, Type: event.Create}
	_, _ = tw.Handler(e, &filenode.ExtraPayload{UUID: uuid.NewString()})
	node := tw.SearchByPath("fs-shadow/test-1")
	assert.NotNil(t, node, "search by name error")

}
