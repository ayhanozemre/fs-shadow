package event

import (
	"fmt"
	connector "github.com/ayhanozemre/fs-shadow/path"
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

type Type string

func (e Type) String() string {
	return string(e)
}

const (
	Remove Type = "remove"
	Write  Type = "write"
	Create Type = "create"
	Rename Type = "rename"
	Move   Type = "move"
)

type Event struct {
	Type     Type
	FromPath connector.Path
	ToPath   connector.Path
}

func (e Event) String() string {
	s := fmt.Sprintf("event %s", e.FromPath.String())
	if e.Type == Rename {
		s += fmt.Sprintf(" -> %s", e.ToPath.String())
	}
	s += fmt.Sprintf(" [%s]", e.Type.String())
	return s
}
