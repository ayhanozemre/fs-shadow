package watcher

import (
	"encoding/json"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	"github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	"sync"
)

type TreeWatcher struct {
	FileTree   *filenode.FileNode
	Watcher    *fsnotify.Watcher
	Path       connector.Path
	ParentPath connector.Path

	Events chan EventTransaction
	Errors chan error

	sync.Mutex
	EventManager event.EventHandler
}

func (tw *TreeWatcher) GetEvents() <-chan EventTransaction {
	return tw.Events
}

func (tw *TreeWatcher) GetErrors() <-chan error {
	return tw.Errors
}

func (tw *TreeWatcher) SearchByPath(path string) *filenode.FileNode {
	return nil
}

func (tw *TreeWatcher) PrintTree(label string) {

	bannerStartLine := fmt.Sprintf("----------------%s----------------", label)
	bannerEndLine := fmt.Sprintf("----------------%s----------------\n\n", label)
	fmt.Println(bannerStartLine)
	a, _ := json.MarshalIndent(tw.FileTree, "", "  ")
	//a, _ := json.Marshal(tw.FileTree)
	fmt.Println(string(a))
	fmt.Println(bannerEndLine)
}

func (tw *TreeWatcher) Remove(path connector.Path) (node *filenode.FileNode, err error) {
	return node, err
}

func (tw *TreeWatcher) Write(path connector.Path) (node *filenode.FileNode, err error) {

	return node, err
}

func (tw *TreeWatcher) Create(path connector.Path, extra *filenode.ExtraPayload) (node *filenode.FileNode, err error) {
	return node, err
}

func (tw *TreeWatcher) Rename(fromPath connector.Path, toPath connector.Path) (node *filenode.FileNode, err error) {
	return node, err
}

func (tw *TreeWatcher) Move(fromPath connector.Path, toPath connector.Path) (node *filenode.FileNode, err error) {
	return node, err
}

func (tw *TreeWatcher) Handler(e event.Event, extras ...*filenode.ExtraPayload) (*EventTransaction, error) {
	return nil, nil
}

func (tw *TreeWatcher) Watch() {

}

func (tw *TreeWatcher) Start() {

}

func (tw *TreeWatcher) Stop() {

}

func (tw *TreeWatcher) Restore(tree *filenode.FileNode) {
}

func NewPathWatcher(fsPath string) (*TreeWatcher, *EventTransaction, error) {
	return nil, nil, nil
}
