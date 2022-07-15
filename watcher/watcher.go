package watcher

import (
	"github.com/ayhanozemre/fs-shadow/event"
	connector "github.com/ayhanozemre/fs-shadow/path"
)

type Watcher interface {
	PrintTree(label string) // sil beni!
	Start()
	Watch()
	Close()
	Handler(event event.Event) error
	Create(fromPath connector.Path) error
	Write(fromPath connector.Path) error
	Rename(fromPath connector.Path, toPath connector.Path) error
	Remove(fromPath connector.Path) error
}

func NewFSWatcher(fsPath string) (Watcher, error) {
	return NewPathWatcher(fsPath)
}

func NewVirtualWatcher(fsPath string) (Watcher, error) {
	return NewVirtualPathWatcher(fsPath)
}
