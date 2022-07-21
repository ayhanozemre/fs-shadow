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

	Events chan event.Event // bu channel'i servislere verecegiz. not implemented
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

	if toPath.IsDir() {
		fmt.Println("dirrrr")
		time.Sleep(1 * time.Second)
		//err1 := tw.Watcher.Add(toPath.String())
		err = tw.Watcher.Remove(fromPath.String())
		if err != nil {
			fmt.Println(toPath.String())
			fmt.Println(fromPath.String())
			fmt.Println("!!!!!!!!!!err", err)
			return err
		}
	}
	return err
}

func (tw *TreeWatcher) Handler(e event.Event) (err error) {
	tw.Lock()
	defer tw.Unlock()
	fromPath := connector.NewFSPath(e.FromPath)

	switch e.Type {
	case event.Remove:
		err = tw.Remove(fromPath)
	case event.Write:
		err = tw.Write(fromPath)
	case event.Create:
		err = tw.Create(fromPath)
	case event.Rename:
		toPath := connector.NewFSPath(e.ToPath)
		err = tw.Rename(fromPath, toPath)
	default:
		errorMsg := fmt.Sprintf("unhandled event: %s", e.String())
		return errors.New(errorMsg)
	}
	return err
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
			fmt.Println("Event->", e)
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
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				if tw.EventManager.StackLength() > 0 {
					newEvents := tw.EventManager.Process()
					for _, e := range newEvents {
						fmt.Println(e.String())
						err := tw.Handler(e)
						if err != nil {
							// event channel update
							fmt.Println("!", err)
						}
					}
				}
				tw.PrintTree("TREE")
				fmt.Println(tw.Watcher.WatchList())

			}
		}
	}()
	go tw.Watch()
}

func NewPathWatcher(fsPath string) (*TreeWatcher, error) {
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
		EventManager: event.NewEventHandler(),
	}
	err = tw.Create(path)
	if err != nil {
		return nil, err
	}
	tw.Start()
	return &tw, nil
}