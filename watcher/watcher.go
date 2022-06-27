package watcher

import (
	"fmt"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"runtime"
)

type Watcher interface {
	PrintTree(label string) // sil beni!
	Start()
	Watch()
	Close()
	EventHandler(op EventType, path string) error
	Create(path connector.Path) error
	Write(path connector.Path) error
	Rename(path connector.Path) error
	Remove(path connector.Path) error
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
