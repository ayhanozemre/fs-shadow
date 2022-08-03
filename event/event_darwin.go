package event

import (
	"github.com/fsnotify/fsnotify"
	"sync"
)

type EventManager struct {
	stack    []fsnotify.Event
	sumStack []string
	sync.Mutex
}

func newEventHandler() *EventManager {
	return &EventManager{stack: []fsnotify.Event{}}
}

func (e *EventManager) Append(event fsnotify.Event, sum string) {

}

func (e *EventManager) StackLength() int {
	return 0
}

func (e *EventManager) Pop() fsnotify.Event {
	return fsnotify.Event{}
}

func (e *EventManager) isCreate(e1, e2, e3, e4, e5, e6 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int) {
	return nil, 0
}

func (e *EventManager) isRemove(e1, e2 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int) {
	return nil, 0
}

func (e *EventManager) isRename(e1, e2, e3, e4, e5 *fsnotify.Event) (*Event, int) {
	return nil, 0
}

func (e *EventManager) isWrite(e1 *fsnotify.Event) (*Event, int) {
	return nil, 0
}

func (e *EventManager) Process() []Event {
	return []Event{}
}
