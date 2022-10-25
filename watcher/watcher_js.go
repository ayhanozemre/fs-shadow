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

type TreeWatcher struct {
	FileTree   *filenode.FileNode
	Path       connector.Path
	ParentPath connector.Path

	sync.Mutex
}

func (tw *TreeWatcher) GetEvents() <-chan EventTransaction {
	return nil
}

func (tw *TreeWatcher) GetErrors() <-chan error {
	return nil
}

func (tw *TreeWatcher) SearchByPath(path string) *filenode.FileNode {
	return nil
}

func (tw *TreeWatcher) PrintTree(label string) {
	bannerStartLine := fmt.Sprintf("----------------%s----------------", label)
	bannerEndLine := fmt.Sprintf("----------------%s----------------\n\n", label)
	fmt.Println(bannerStartLine)
	a, _ := json.MarshalIndent(tw.FileTree, "", "  ")
	fmt.Println(string(a))
	fmt.Println(bannerEndLine)
}

func (tw *TreeWatcher) Create(path connector.Path, extra *filenode.ExtraPayload) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *TreeWatcher) Remove(path connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *TreeWatcher) Rename(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *TreeWatcher) Move(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *TreeWatcher) Write(path connector.Path) (*filenode.FileNode, error) {
	return nil, nil
}

func (tw *TreeWatcher) Stop() {
	log.Debug("close not implemented ")
}

func (tw *TreeWatcher) Start() {
	log.Debug("start not implemented ")
}

func (tw *TreeWatcher) Watch() {
	log.Debug("watch not implemented ")
}

func (tw *TreeWatcher) Handler(e event.Event, extras ...*filenode.ExtraPayload) (*EventTransaction, error) {
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

func (tw *TreeWatcher) Restore(tree *filenode.FileNode) {
	tw.FileTree = tree
}

func NewPathWatcher(virtualPath string, extra *filenode.ExtraPayload) (*TreeWatcher, *EventTransaction, error) {
	return nil, nil, nil
}
