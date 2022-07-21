package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/event"
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

func (tw *VirtualTree) Create(path connector.Path, extra *filenode.ExtraPayload) (*filenode.FileNode, error) {
	eventPath := path.ExcludePath(tw.ParentPath)
	node, err := tw.FileTree.Create(eventPath, path)
	if err != nil {
		return nil, err
	}
	node.UpdateWithExtra(*extra)
	return node, nil
}

func (tw *VirtualTree) Remove(path connector.Path) (*filenode.FileNode, error) {
	eventPath := path.ExcludePath(tw.ParentPath)
	node, err := tw.FileTree.Remove(eventPath)
	return node, err
}

func (tw *VirtualTree) Rename(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error) {
	node, err := tw.FileTree.Rename(fromPath.ExcludePath(tw.ParentPath), toPath.ExcludePath(tw.ParentPath))
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (tw *VirtualTree) Write(path connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *VirtualTree) Close() {
	panic("close not implemented ")
}

func (tw *VirtualTree) Start() {
	panic("start not implemented ")
}

func (tw *VirtualTree) Watch() {
	panic("watch not implemented ")
}

func (tw *VirtualTree) Handler(e event.Event, extra *filenode.ExtraPayload) (*EventTransaction, error) {
	var err error
	var node *filenode.FileNode
	tw.Lock()
	defer tw.Unlock()

	switch e.Type {
	case event.Remove:
		node, err = tw.Remove(e.FromPath)
	case event.Write:
		node, err = tw.Write(e.FromPath)
	case event.Create:
		node, err = tw.Create(e.FromPath, extra)
	case event.Rename:
		node, err = tw.Rename(e.FromPath, e.ToPath)
	default:
		errorMsg := fmt.Sprintf("unhandled event: %s", e.String())
		err = errors.New(errorMsg)
	}
	if err != nil {
		return nil, err
	}
	et := makeEventTransaction(*node, e.Type)
	return et, err
}

func (tw *VirtualTree) Restore(tree *filenode.FileNode) {
	tw.FileTree = tree
}

func NewVirtualPathWatcher(virtualPath string, extra *filenode.ExtraPayload) (*VirtualTree, *EventTransaction, error) {
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
	e := event.Event{FromPath: path, Type: event.Create}

	txn, err := tw.Handler(e, extra)
	if err != nil {
		return nil, nil, err
	}
	return &tw, txn, nil
}
