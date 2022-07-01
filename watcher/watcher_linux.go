package watcher

import (
	"encoding/json"
	"errors"
	"fmt"
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

	Events chan Event // bu channel'i servislere verecegiz. not implemented
	Errors chan error

	sync.Mutex
	EventManager *EventManager
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
}

func (tw *TreeWatcher) Remove(path connector.Path) error {
	eventPath := path.ExcludePath(tw.ParentPath)
	err, node := tw.FileTree.Remove(eventPath)
	if err == nil && node != nil && node.Meta.IsDir {
		err = tw.Watcher.Remove(path.String())
		if err != nil {
			return err
		}
	}
	return err
}

func (tw *TreeWatcher) Write(path connector.Path) error {
	if !path.IsDir() {
		eventPath := path.ExcludePath(tw.ParentPath)
		err := tw.FileTree.Update(eventPath, path)
		return err
	}
	return nil
}

func (tw *TreeWatcher) Create(path connector.Path) error {
	if !path.Exists() {
		return errors.New("file path does not exist")
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

	err := tw.FileTree.Create(eventPath, path, eventCh)
	eventCh <- nil
	close(eventCh)
	return err
}

func (tw *TreeWatcher) Rename(fromPath connector.Path, toPath connector.Path) error {
	var err error
	err = tw.FileTree.Rename(fromPath.ExcludePath(tw.ParentPath), toPath.ExcludePath(tw.ParentPath))
	if err != nil {
		return err
	}

	err = tw.Watcher.Remove(fromPath.String())
	if err != nil {
		return err
	}

	err = tw.Watcher.Add(toPath.String())
	return err
}

func (tw *TreeWatcher) EventHandler(event Event) (err error) {
	tw.Lock()
	defer tw.Unlock()
	fromPath := connector.NewFSPath(event.FromPath)

	switch event.Type {
	case Remove:
		err = tw.Remove(fromPath)
	case Write:
		err = tw.Write(fromPath)
	case Create:
		err = tw.Create(fromPath)
	case Rename:
		toPath := connector.NewFSPath(event.ToPath)
		err = tw.Rename(fromPath, toPath)
	default:
		errorMsg := fmt.Sprintf("unhandled event: op:%s, path:%s", event.Type, event.FromPath)
		return errors.New(errorMsg)
	}
	return err
}

func (tw *TreeWatcher) Watch() {
	for {
		select {
		case event, ok := <-tw.Watcher.Events:
			if !ok {
				return
			}
			var sum string
			path := connector.NewFSPath(event.Name)
			eventPath := path.ExcludePath(tw.ParentPath)
			node := tw.FileTree.Search(eventPath.ParentPath().String())
			if node != nil {
				sum = node.Meta.Sum
			}
			tw.EventManager.Append(event, sum)
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
				if len(tw.EventManager.stack) > 0 {
					newEvents := tw.EventManager.Process()
					for _, event := range newEvents {
						err := tw.EventHandler(event)
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

func newLinuxPathWatcher(fsPath string) (*TreeWatcher, error) {
	var err error
	var watcher *fsnotify.Watcher
	path := connector.NewFSPath(fsPath)
	if !path.IsDir() {
		err = errors.New("input path is not directory")
		return nil, err
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
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
		EventManager: NewEventHandler(),
	}
	err = tw.Create(path)
	if err != nil {
		return nil, err
	}
	tw.Start()
	return &tw, nil
}
