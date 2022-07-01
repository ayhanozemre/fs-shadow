package watcher

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"sync"
)

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

type EventManager struct {
	stack    []fsnotify.Event
	sumStack []string
	sync.Mutex
}

func NewEventHandler() *EventManager {
	return &EventManager{stack: []fsnotify.Event{}}
}

func (e *EventManager) Append(event fsnotify.Event, sum string) {
	e.Lock()
	e.stack = append(e.stack, event)
	e.sumStack = append(e.sumStack, sum)
	e.Unlock()
}

func (e *EventManager) Pop() fsnotify.Event {
	e.Lock()
	var event fsnotify.Event
	event, e.stack = e.stack[0], e.stack[1:]
	e.Unlock()
	return event
}

func (e *EventManager) isCreate(e1, e2, e3 *fsnotify.Event, e1Sum, e2Sum string) (*Event, int) {
	_, e1FileErr := os.Stat(e1.Name)

	if e1.Op == fsnotify.Create {
		if e2 == nil {
			fmt.Println("create1")
			return &Event{FromPath: e1.Name, Type: Create}, 1
		} else if e2.Op == fsnotify.Chmod && e1.Name == e2.Name {
			fmt.Println("create2")
			return &Event{FromPath: e1.Name, Type: Create}, 2
		} else if e2.Op == fsnotify.Rename && e1.Name == e2.Name && e1Sum == e2Sum && e1FileErr == nil {
			fmt.Println("create3")
			return &Event{FromPath: e1.Name, Type: Create}, 2
		} else {
			fmt.Println("create4")
			return &Event{FromPath: e1.Name, Type: Create}, 1
		}
	}
	return nil, 0
}

func (e *EventManager) isRemove(e1, e2, e3 *fsnotify.Event, e1Sum, e2Sum, e3Sum string) (*Event, int) {
	_, e1FileErr := os.Stat(e1.Name)

	if e1.Op == fsnotify.Remove && e2 != nil && e2.Op == fsnotify.Remove && e1.Name == e2.Name {
		// watcher'a eklenmis bir klasoru sildigimizde bu case gerceklesecek.
		fmt.Println("remove1")
		return &Event{FromPath: e1.Name, Type: Remove}, 2
	}

	if e1.Op == fsnotify.Rename && e2 != nil && e2.Op == fsnotify.Rename && e1.Name == e2.Name && os.IsNotExist(e1FileErr) {
		// watcher'a eklenmis bir klasoru watch etmedigimiz bir klasore tasirsak bu case gerceklesecek
		fmt.Println("remove2")
		return &Event{FromPath: e1.Name, Type: Remove}, 2
	}
	//fmt.Println("e1/e2 sum", e1Sum, e2Sum)
	if e1.Op == fsnotify.Rename && e2 != nil && e2.Op == fsnotify.Create && e1Sum != e2Sum {
		fmt.Println("extremeRemove")
		return &Event{FromPath: e1.Name, Type: Remove}, 1
	}

	if e1.Op == fsnotify.Rename && e2 == nil && os.IsNotExist(e1FileErr) {
		fmt.Println("remove4")
		return &Event{FromPath: e1.Name, Type: Remove}, 1
	}

	if e1.Op == fsnotify.Remove {
		fmt.Println("remove5")
		return &Event{FromPath: e1.Name, Type: Remove}, 1
	}
	return nil, 0
}

func (e *EventManager) isRename(e1, e2, e3 *fsnotify.Event) (*Event, int) {
	if e1.Op == fsnotify.Rename {
		if e2 != nil && e2.Op == fsnotify.Create && e3 != nil && e3.Op == fsnotify.Rename && e1.Name == e3.Name {
			// watcher'a eklenmis bir klasoru rename yaptigimizda bu case gerceklesecek.
			fmt.Println("rename1")
			return &Event{FromPath: e1.Name, ToPath: e2.Name, Type: Rename}, 3
		} else if e2 != nil && e2.Op == fsnotify.Create && e1.Name != e2.Name {
			// watcher'a eklenmemis bir klasoru rename yaptigimizda bu case gerceklesecek.
			fmt.Println("rename2")
			return &Event{FromPath: e1.Name, ToPath: e2.Name, Type: Rename}, 2
		}
	}
	return nil, 0
}

func (e *EventManager) isWrite(e1 *fsnotify.Event) (*Event, int) {
	if e1.Op == fsnotify.Write {
		return &Event{FromPath: e1.Name, Type: Rename}, 1
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
		var e1, e2, e3 *fsnotify.Event
		var e1Sum, e2Sum, e3Sum string
		if cursor >= sl {
			break
		}
		e1 = &e.stack[cursor]
		e1Sum = e.sumStack[cursor]
		if cursor+1 < sl {
			e2 = &e.stack[cursor+1]
			e2Sum = e.sumStack[cursor+1]
		}
		if cursor+2 < sl {
			e3 = &e.stack[cursor+2]
			e3Sum = e.sumStack[cursor+2]
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
		if event, nc := e.isRemove(e1, e2, e3, e1Sum, e2Sum, e3Sum); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			fmt.Println(event.String())
			continue
		}
		if event, nc := e.isCreate(e1, e2, e3, e1Sum, e2Sum); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			fmt.Println(event.String())
			// break and generate sum for nodeTree
			break
		}
		if event, nc := e.isRename(e1, e2, e3); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			fmt.Println(event.String())
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
