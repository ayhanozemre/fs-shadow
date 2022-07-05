package watcher

import (
	"github.com/ayhanozemre/fs-shadow/event"
	connector "github.com/ayhanozemre/fs-shadow/path"
	log "github.com/sirupsen/logrus"
	"runtime"
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
	var watcher Watcher
	var err error
	switch os := runtime.GOOS; os {
	case "darwin":
		log.Debug("OS X watcher not implemented")
	case "windows":
		log.Debug("windows watcher not implemented")
	default:
		watcher, err = NewLinuxPathWatcher(fsPath)
	}
	return watcher, err
}
