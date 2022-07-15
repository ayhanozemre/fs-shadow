package event

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
)

type EventHandler interface {
	StackLength() int
	Append(event fsnotify.Event, sum string)
	Pop() fsnotify.Event
	isCreate(e1, e2, e3, e4, e5, e6 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int)
	isRemove(e1, e2 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int)
	isRename(e1, e2, e3, e4, e5 *fsnotify.Event) (*Event, int)
	isWrite(e1 *fsnotify.Event) (*Event, int)
	Process() []Event
}

func NewEventHandler() EventHandler {
	return newEventHandler()
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
