package watcher

import (
	"fmt"
	"github.com/ayhanozemre/fs-shadow/event"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"runtime"
)

type Watcher interface {
	PrintTree(label string) // sil beni!
	Start()
	Watch()
	Close()
	EventHandler(event event.Event) error
	Create(fromPath connector.Path) error
	Write(fromPath connector.Path) error
	Rename(fromPath connector.Path, toPath connector.Path) error
	Remove(fromPath connector.Path) error
}

func NewFSWatcher(fsPath string) (Watcher, error) {
	var watcher Watcher
	var err error
	switch os := runtime.GOOS; os {
	case "darwin":
		fmt.Println("OS X watcher")
	case "windows":
		fmt.Println("windows watcher")
	default:
		watcher, err = newLinuxPathWatcher(fsPath)
	}
	return watcher, err
}
