package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/filenode"
	"github.com/ayhanozemre/fs-shadow/path"
	"sync"
)

type VirtualTree struct {
	FileTree   *filenode.FileNode
	Path       connector.Path
	ParentPath connector.Path

	sync.Mutex
}

func (tw *VirtualTree) PrintTree(label string) {
	bannerStartLine := fmt.Sprintf("----------------%s-----------------------", label)
	bannerEndLine := fmt.Sprintf("----------------%s-----------------------\n\n", label)
	fmt.Println(bannerStartLine)
	a, _ := json.Marshal(tw.FileTree)
	fmt.Println(string(a))
	fmt.Println(bannerEndLine)
}

func (tw *VirtualTree) Create(path connector.Path) error {
	tw.Lock()
	defer tw.Unlock()
	if !path.Exists() {
		return errors.New("file path does not exist")
	}

	eventPath := path.ExcludePath(tw.ParentPath)
	eventCh := make(chan connector.Path)
	go func() {
		for {
			select {
			case p := <-eventCh:
				if p != nil {
					//
				}
			}
		}
	}()
	err := tw.FileTree.Create(eventPath, path, eventCh)
	close(eventCh)
	return err
}

func (tw *VirtualTree) Remove(path connector.Path) error {
	tw.Lock()
	defer tw.Unlock()
	eventPath := path.ExcludePath(tw.ParentPath)
	err, _ := tw.FileTree.Remove(eventPath)
	return err
}

func (tw *VirtualTree) Write(path connector.Path) error {
	// maybe we can generate the sum for vfs
	return nil
}

func (tw *VirtualTree) Close() {
}

func (tw *VirtualTree) Rename(fromPath connector.Path, toPath connector.Path) error {
	tw.Lock()
	defer tw.Unlock()
	var err error
	err = tw.FileTree.Rename(fromPath.ExcludePath(tw.ParentPath), toPath.ExcludePath(tw.ParentPath))
	if err != nil {
		return err
	}
	return nil
}

func NewVirtualPathWatcher(virtualPath string) (*VirtualTree, error) {
	path := connector.NewVirtualPath(virtualPath, true)

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
	err := tw.Create(path)
	if err != nil {
		return nil, err
	}
	return &tw, nil
}
