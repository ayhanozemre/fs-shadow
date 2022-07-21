package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	"github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type TreeWatcher struct {
	FileTree   *filenode.FileNode
	Watcher    *fsnotify.Watcher
	Path       connector.Path
	ParentPath connector.Path

	Events chan EventTransaction // bu channel'i servislere verecegiz. not implemented
	Errors chan error

	sync.Mutex
	EventManager event.EventHandler
}

func (tw *TreeWatcher) PrintTree(label string) {
	bannerStartLine := fmt.Sprintf("----------------%s-----------------------", label)
	bannerEndLine := fmt.Sprintf("----------------%s-----------------------\n\n", label)
	fmt.Println(bannerStartLine)
	a, _ := json.Marshal(tw.FileTree)
	fmt.Println(string(a))
	fmt.Println(bannerEndLine)
}

func (tw *TreeWatcher) Close() {
	err := tw.Watcher.Close()
	if err != nil {
		log.Error(err)
	}
	close(tw.Events)
	close(tw.Errors)
}

func (tw *TreeWatcher) Remove(path connector.Path) (*filenode.FileNode, error) {
	eventPath := path.ExcludePath(tw.ParentPath)
	node, err := tw.FileTree.Remove(eventPath)
	if err == nil && node != nil && node.Meta.IsDir {
		err = tw.Watcher.Remove(path.String())
		if err != nil {
			return nil, err
		}
	}
	return node, err
}

func (tw *TreeWatcher) Write(path connector.Path) (*filenode.FileNode, error) {
	var node *filenode.FileNode
	var err error
	if !path.IsDir() {
		eventPath := path.ExcludePath(tw.ParentPath)
		node, err = tw.FileTree.Update(eventPath, path)
		return nil, err
	}
	return node, err
}

func (tw *TreeWatcher) Create(path connector.Path, extra *filenode.ExtraPayload) (*filenode.FileNode, error) {
	if !path.Exists() {
		return nil, errors.New("file path does not exist")
	}

	eventPath := path.ExcludePath(tw.ParentPath)
	eventCh := make(chan connector.Path)

	go func() {
		for {
			select {
			case p := <-eventCh:
				if p != nil {
					if p.IsDir() {
						err := tw.Watcher.Add(p.String())
						if err != nil {
							fmt.Println("create error", err)
							return
						}
					}
				} else {
					return
				}
			}
		}
	}()

	node, err := tw.FileTree.Create(eventPath, path, eventCh)
	eventCh <- nil
	close(eventCh)
	return node, err
}

func (tw *TreeWatcher) Rename(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error) {
	node, err := tw.FileTree.Rename(fromPath.ExcludePath(tw.ParentPath), toPath.ExcludePath(tw.ParentPath))
	if err != nil {
		return nil, err
	}
	if node.Meta.IsDir {
		err = tw.Watcher.Remove(fromPath.String())
		if err != nil {
			return nil, err
		}

		err = tw.Watcher.Add(toPath.String())
	}
	return node, err
}

func (tw *TreeWatcher) Handler(e event.Event, extra *filenode.ExtraPayload) (*EventTransaction, error) {
	tw.Lock()
	var err error
	var node *filenode.FileNode
	defer tw.Unlock()

	switch e.Type {
	case event.Remove:
		node, err = tw.Remove(e.FromPath)
	case event.Write:
		node, err = tw.Write(e.FromPath)
	case event.Create:
		node, err = tw.Create(e.FromPath, extra)
	case event.Rename:
		node, err = tw.Rename(e.FromPath, e.ToPath)
	default:
		errorMsg := fmt.Sprintf("unhandled event: op:%s, path:%s", e.Type, e.FromPath)
		err = errors.New(errorMsg)
	}
	if err != nil {
		return nil, err
	}
	et := makeEventTransaction(*node, e.Type)
	return et, err
}

func (tw *TreeWatcher) Watch() {
	for {
		select {
		case e, ok := <-tw.Watcher.Events:
			if !ok {
				return
			}
			var sum string
			path := connector.NewFSPath(e.Name)
			eventPath := path.ExcludePath(tw.ParentPath)
			node := tw.FileTree.Search(eventPath.ParentPath().String())
			if node != nil {
				sum = node.Meta.Sum
			}
			tw.EventManager.Append(e, sum)
		case err, ok := <-tw.Watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("error:", err)
		}
	}
}

func (tw *TreeWatcher) Start() {
	fmt.Println("started!")
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				if tw.EventManager.StackLength() > 0 {
					newEvents := tw.EventManager.Process()
					for _, e := range newEvents {
						_, err := tw.Handler(e, nil)
						if err != nil {
							// event channel update
							fmt.Println(err)
						}
					}
				}
			}
		}
	}()
	go tw.Watch()
}

func (tw *TreeWatcher) Restore(tree *filenode.FileNode) {
	tw.FileTree = tree
}

func NewPathWatcher(fsPath string) (*TreeWatcher, *EventTransaction, error) {
	var err error
	var watcher *fsnotify.Watcher
	path := connector.NewFSPath(fsPath)
	if !path.IsDir() {
		err = errors.New("input path is not directory")
		return nil, nil, err
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}

	root := filenode.FileNode{
		Name: path.Name(),
		Meta: filenode.MetaData{
			IsDir: true,
		},
	}

	tw := TreeWatcher{
		FileTree:     &root,
		ParentPath:   path.ParentPath(),
		Path:         path,
		Watcher:      watcher,
		EventManager: event.NewEventHandler(),
		Events:       make(chan EventTransaction),
		Errors:       make(chan error),
	}
	e := event.Event{FromPath: path, Type: event.Create}
	txn, err := tw.Handler(e, nil)
	if err != nil {
		return nil, nil, err
	}
	tw.Start()
	return &tw, txn, nil
}
