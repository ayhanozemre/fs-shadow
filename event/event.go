package event

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"runtime"
)

type EventHandler interface {
	StackLength() int
	Append(event fsnotify.Event, sum string)
	Pop() fsnotify.Event
	isCreate(e1, e2 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int)
	isRemove(e1, e2 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int)
	isRename(e1, e2, e3 *fsnotify.Event) (*Event, int)
	isWrite(e1 *fsnotify.Event) (*Event, int)
	Process() []Event
}

func NewEventHandler() EventHandler {
	var handler EventHandler
	switch os := runtime.GOOS; os {
	case "darwin":
		log.Debug("OS X event handler not implemented")
	case "windows":
		log.Debug("windows event handler not implemented")
	default:
		handler = newLinuxEventHandler()
	}
	return handler
}

type EventType string

func (e EventType) String() string {
	return string(e)
}

const (
	Remove EventType = "remove"
	Write  EventType = "write"
	Create EventType = "create"
	Rename EventType = "rename"
)

type Event struct {
	Type     EventType
	FromPath string
	ToPath   string
}

func (e Event) String() string {
	s := fmt.Sprintf("rsult %s", e.FromPath)
	if e.Type == Rename {
		s += fmt.Sprintf(" -> %s", e.ToPath)
	}
	s += fmt.Sprintf(" [%s]", e.Type)
	return s
}
