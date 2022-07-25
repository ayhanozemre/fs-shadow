package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	"github.com/ayhanozemre/fs-shadow/path"
	log "github.com/sirupsen/logrus"
	"sync"
)

type VirtualTree struct {
	FileTree   *filenode.FileNode
	Path       connector.Path
	ParentPath connector.Path

	sync.Mutex
}

func (tw *VirtualTree) GetEvents() <-chan EventTransaction {
	return nil
}

func (tw *VirtualTree) GetErrors() <-chan error {
	return nil
}

func (tw *VirtualTree) PrintTree(label string) {
	bannerStartLine := fmt.Sprintf("----------------%s----------------", label)
	bannerEndLine := fmt.Sprintf("----------------%s----------------\n\n", label)
	fmt.Println(bannerStartLine)
	a, _ := json.MarshalIndent(tw.FileTree, "", "  ")
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
	log.Debug("close not implemented ")
}

func (tw *VirtualTree) Start() {
	log.Debug("start not implemented ")
}

func (tw *VirtualTree) Watch() {
	log.Debug("watch not implemented ")
}

func (tw *VirtualTree) Handler(e event.Event, extras ...*filenode.ExtraPayload) (*EventTransaction, error) {
	tw.Lock()
	defer tw.Unlock()
	var err error
	var node *filenode.FileNode
	var extra *filenode.ExtraPayload

	if len(extras) > 0 {
		extra = extras[0]
	}

	switch e.Type {
	case event.Remove:
		node, err = tw.Remove(e.FromPath)
		break
	case event.Write:
		node, err = tw.Write(e.FromPath)
		break
	case event.Create:
		node, err = tw.Create(e.FromPath, extra)
		break
	case event.Rename:
		node, err = tw.Rename(e.FromPath, e.ToPath)
		break
	default:
		errorMsg := fmt.Sprintf("unhandled event: %s", e.String())
		err = errors.New(errorMsg)
		break
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
