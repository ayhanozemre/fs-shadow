package event

import (
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
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

func (e *EventManager) isCreate(e1, e2, e3, e4, e5, e6 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int) {
	if e1.Op == fsnotify.Create && e2 != nil && e2.Op == fsnotify.Remove|fsnotify.Rename {
		log.Debug("create-case-1")
		return nil, 0
	}
	if e1.Op == fsnotify.Create && e2 != nil && e2.Op == fsnotify.Rename {
		log.Debug("create-case-2")
		return nil, 0
	}
	if e1.Op == fsnotify.Create {
		log.Debug("create-case-3")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Create}, 1
	}

	return nil, 0
}

func (e *EventManager) isRemove(e1, e2 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int) {

	if e1.Op == fsnotify.Remove {
		log.Debug("remove-case-1")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Remove}, 1
	}
	if e1.Op == fsnotify.Remove|fsnotify.Rename {
		log.Debug("remove-case-2")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Remove}, 1
	}

	if e1.Op == fsnotify.Remove|fsnotify.Write {
		log.Debug("remove-case-3")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Remove}, 1
	}

	_, e1FileErr := os.Stat(e1.Name)
	if e1.Op == fsnotify.Rename && os.IsNotExist(e1FileErr) {
		// move to outside
		log.Debug("remove-case-4")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Remove}, 1
	}

	return nil, 0
}

func (e *EventManager) isRename(e1, e2, e3, e4, e5 *fsnotify.Event) (*Event, int) {
	if e1.Op == fsnotify.Create && e2 != nil && e2.Op == fsnotify.Remove|fsnotify.Rename {
		log.Debug("rename-case-1")
		return &Event{FromPath: connector.NewFSPath(e2.Name), ToPath: connector.NewFSPath(e1.Name), Type: Rename}, 2
	}
	if e1.Op == fsnotify.Create && e2 != nil && e2.Op == fsnotify.Rename {
		log.Debug("rename-case-2")
		return &Event{FromPath: connector.NewFSPath(e2.Name), ToPath: connector.NewFSPath(e1.Name), Type: Rename}, 2
	}
	return nil, 0
}

func (e *EventManager) isWrite(e1 *fsnotify.Event) (*Event, int) {
	if e1.Op == fsnotify.Write {
		log.Debug("write-case-1")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Rename}, 1
	}
	return nil, 0
}

func (e *EventManager) Process() []Event {
	e.Lock()
	defer e.Unlock()
	cursor := 0
	sl := len(e.stack)
	var newEvents []Event
	for {
		var e1, e2, e3, e4, e5, e6 *fsnotify.Event
		var e1Sum, e2Sum string
		if cursor >= sl {
			break
		}

		e1 = &e.stack[cursor]
		e1Sum = e.sumStack[cursor]
		if cursor+1 < sl {
			e2 = &e.stack[cursor+1]
			e2Sum = e.sumStack[cursor+1]
		}

		if e1.Op == fsnotify.Chmod {
			cursor += 1
			continue
		}

		if event, nc := e.isWrite(e1); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			continue
		}
		if event, nc := e.isRemove(e1, e2, e1Sum, e2Sum); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			log.Debug(event.String())
			continue
		}
		if event, nc := e.isCreate(e1, e2, e3, e4, e5, e6, e1Sum, e2Sum); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			log.Debug(event.String())
			// break and generate sum for nodeTree
			break
		}
		if event, nc := e.isRename(e1, e2, e3, e4, e5); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			log.Debug(event.String())
			continue
		}
		break
	}
	if cursor == sl {
		e.stack = []fsnotify.Event{}
		e.sumStack = []string{}
	} else {
		e.stack = e.stack[sl-(sl-cursor):]
		e.sumStack = e.sumStack[sl-(sl-cursor):]
	}
	return newEvents
}
