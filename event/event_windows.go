package event

import (
	"fmt"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime/debug"
	"sync"
	"time"
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
	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			fmt.Println(os.Stderr, "1Exception: %v\n", err)
			fmt.Println(string(debug.Stack()))

			time.Sleep(20 * time.Second)
		}
	}()
	//_, e1FileErr := os.Stat(e1.Name)

	//_, e5FileErr := os.Stat(e5.Name)

	if e1.Op == fsnotify.Create &&
		e2 != nil && e2.Op == fsnotify.Write &&
		e3 != nil && e3.Op == fsnotify.Rename &&
		e4 != nil && e4.Op == fsnotify.Create &&
		e5 != nil && e5.Op == fsnotify.Write &&
		e6 != nil && e6.Op == fsnotify.Write &&
		e1.Name == e3.Name && e4.Name == e6.Name {
		log.Debug("extreme create")
		return &Event{FromPath: connector.NewFSPath(e6.Name), Type: Create}, 6
	}

	if e1.Op == fsnotify.Create &&
		e2 != nil && e2.Op == fsnotify.Write &&
		e3 != nil && e3.Op == fsnotify.Rename &&
		e4 != nil && e4.Op == fsnotify.Create &&
		e5 != nil && e5.Op == fsnotify.Write &&
		e1.Name == e3.Name {
		log.Debug("extreme create22")
		return &Event{FromPath: connector.NewFSPath(e4.Name), Type: Create}, 5
	}

	if e1.Op == fsnotify.Create {
		if e2 == nil {
			log.Debug("create1")
			return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Create}, 1
		} else if e2.Op == fsnotify.Rename && e1Sum == e2Sum &&
			e3 != nil && e3.Op == fsnotify.Create &&
			e4 != nil && e4.Op == fsnotify.Write &&
			e3.Name == e4.Name {
			log.Debug("create3xxx")
			return &Event{FromPath: connector.NewFSPath(e4.Name), Type: Create}, 4

		} else if e2.Op == fsnotify.Write && e1.Name == e2.Name {
			log.Debug("create3")
			return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Create}, 2
		} else if e2.Op == fsnotify.Rename && e1.Name == e2.Name {
			log.Debug("create2")
			return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Create}, 2
		} else {
			log.Debug("create6")
			return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Create}, 1
		}
		/*
			else if e2.Op == fsnotify.Write && os.IsNotExist(e1FileErr) {
					log.Debug("create4")
					return &Event{FromPath: e1.Name, Type: Create}, 2
				} */
	}
	return nil, 0
}

func (e *EventManager) isRemove(e1, e2 *fsnotify.Event, e1sum, e2sum string) (*Event, int) {
	if e1.Op == fsnotify.Remove && e2 != nil && e2.Op == fsnotify.Remove && e1.Name == e2.Name {
		// watcher'a eklenmis bir klasoru sildigimizde bu case gerceklesecek.
		log.Debug("remove1")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Remove}, 2
	}

	if e1.Op == fsnotify.Remove {
		log.Debug("remove2")
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Remove}, 1
	}
	return nil, 0
}

func (e *EventManager) isRename(e1, e2, e3, e4, e5 *fsnotify.Event) (*Event, int) {
	if e1.Op == fsnotify.Rename {
		if e2 != nil && e2.Op == fsnotify.Create && e3 != nil && e3.Op == fsnotify.Write &&
			e4 != nil && e4.Op == fsnotify.Write && e5 != nil && e5.Op == fsnotify.Write &&
			e2.Name == e4.Name && e2.Name == e5.Name {
			log.Debug("rename9")
			return &Event{FromPath: connector.NewFSPath(e1.Name), ToPath: connector.NewFSPath(e5.Name), Type: Rename}, 5
		}

		if e2 != nil && e2.Op == fsnotify.Create && e3 != nil && e3.Op == fsnotify.Write && e2.Name == e3.Name {

			if e4 != nil && e4.Op == fsnotify.Write && e3.Name == e4.Name {
				log.Debug("rename111")
				return &Event{FromPath: connector.NewFSPath(e1.Name), ToPath: connector.NewFSPath(e2.Name), Type: Rename}, 4
			} else {
				log.Debug("rename1")
				return &Event{FromPath: connector.NewFSPath(e1.Name), ToPath: connector.NewFSPath(e2.Name), Type: Rename}, 3
			}
		} else if e2 != nil && e2.Op == fsnotify.Create && e1.Name != e2.Name {
			log.Debug("rename2")
			return &Event{FromPath: connector.NewFSPath(e1.Name), ToPath: connector.NewFSPath(e2.Name), Type: Rename}, 2
		}
	}
	return nil, 0
}

func (e *EventManager) isWrite(e1 *fsnotify.Event) (*Event, int) {
	if e1.Op == fsnotify.Write {
		return &Event{FromPath: connector.NewFSPath(e1.Name), Type: Write}, 1
	}
	return nil, 0
}

func (e *EventManager) Process() []Event {
	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			fmt.Println(os.Stderr, "Exception: %v\n", err)
			fmt.Println(string(debug.Stack()))
			time.Sleep(20 * time.Second)
		}
	}()
	e.Lock()
	defer e.Unlock()
	cursor := 0
	sl := len(e.stack)
	var newEvents []Event
	for {
		// Normal sartlar altinda butun isletim sistemlerinde donen event sayisi 3tur.
		// Fakat windows isletim sisteminde bu sayi 6 olarak gozlendi. Bundan dolayi
		// 6 adet event'e gore event bus olusturuldu.
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
		if cursor+2 < sl {
			e3 = &e.stack[cursor+2]
		}
		if cursor+3 < sl {
			e4 = &e.stack[cursor+3]
		}
		if cursor+4 < sl {
			e5 = &e.stack[cursor+4]
		}
		if cursor+5 < sl {
			e6 = &e.stack[cursor+5]
		}

		if e1.Op == fsnotify.Chmod {
			cursor += 1
			continue
		}

		e1Path := connector.NewFSPath(e1.Name)
		if e1.Op == fsnotify.Write && e1Path.IsDir() {
			cursor += 1
			continue
		}

		fmt.Println("------------------------")
		fmt.Println("E1-", e1, cursor, sl)
		fmt.Println("E2-", e2)
		fmt.Println("E3-", e3)
		fmt.Println("E4-", e4)
		fmt.Println("E5-", e5)
		fmt.Println("------------------------")

		if event, nc := e.isWrite(e1); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			continue
		}
		if event, nc := e.isRemove(e1, e2, e1Sum, e2Sum); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			fmt.Println(event.String())
			continue
		}
		if event, nc := e.isCreate(e1, e2, e3, e4, e5, e6, e1Sum, e2Sum); event != nil {
			cursor += nc
			newEvents = append(newEvents, *event)
			fmt.Println(event.String())
			// break and generate sum for nodeTree
			break
		}
		if event, nc := e.isRename(e1, e2, e3, e4, e5); event != nil {
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
