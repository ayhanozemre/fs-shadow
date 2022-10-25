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

func (tw *VirtualTree) SearchByPath(path string) *filenode.FileNode {
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
	return nil, nil
}

func (tw *VirtualTree) Remove(path connector.Path) (*filenode.FileNode, error) {
	return nil, err
}

func (tw *VirtualTree) Rename(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *VirtualTree) Move(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *VirtualTree) Write(path connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *VirtualTree) Stop() {
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
	case event.Move:
		node, err = tw.Move(e.FromPath, e.ToPath)
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
	return nil, nil, nil
}
