package event

import (
	"github.com/fsnotify/fsnotify"
	"sync"
)

/*
	Since OS events arrive synchronously and some events alone do not make sense, there was a need to set up a queue structure.
	Watchers written for OS periodically push events to the EventManager's stack and periodically handle these events within the Process method.
*/
type EventManager struct {
	stack    []fsnotify.Event
	sumStack []string
	sync.Mutex
}

func newEventHandler() *EventManager {
	return &EventManager{stack: []fsnotify.Event{}}
}

func (e *EventManager) Append(event fsnotify.Event, sum string) {
	e.Lock()
	e.stack = append(e.stack, event)
	e.sumStack = append(e.sumStack, sum)
	e.Unlock()
}

func (e *EventManager) StackLength() int {
	return len(e.stack)
}

func (e *EventManager) Pop() fsnotify.Event {
	e.Lock()
	var event fsnotify.Event
	event, e.stack = e.stack[0], e.stack[1:]
	e.Unlock()
	return event
}

func (e *EventManager) isCreate(e1, e2, _, _, _, _ *fsnotify.Event, e1Sum, e2Sum string) (*Event, int) {

	return nil, 0
}

func (e *EventManager) isRemove(e1, e2 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int) {
	return nil, 0
}

func (e *EventManager) isRename(e1, e2, e3, _, _ *fsnotify.Event) (*Event, int) {
	return nil, 0
}

func (e *EventManager) isWrite(e1 *fsnotify.Event) (*Event, int) {
	return nil, 0
}

// This is where Event Manager processes the events in the main bus.Stack piece by piece.
// Determining the maximum number of events to be processed;
//  It is determined by how many events the operating system will send for a file transaction.
func (e *EventManager) Process() []Event {
	return newEvents
}
