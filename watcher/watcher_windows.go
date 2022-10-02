package watcher

import (
	"context"
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

	Events chan EventTransaction
	Errors chan error

	IgniterReloadFunc func()
	IgniterReloadCtx  context.Context

	sync.Mutex
	EventManager event.EventHandler
}

func (tw *TreeWatcher) GetEvents() <-chan EventTransaction {
	return tw.Events
}

func (tw *TreeWatcher) GetErrors() <-chan error {
	return tw.Errors
}

func (tw *TreeWatcher) Restore(tree *filenode.FileNode) {
	tw.FileTree = tree
}

func (tw *TreeWatcher) SearchByPath(path string) *filenode.FileNode {
	return tw.FileTree.Search(path)
}

func (tw *TreeWatcher) PrintTree(label string) {
	bannerStartLine := fmt.Sprintf("----------------%s-----------------------", label)
	bannerEndLine := fmt.Sprintf("----------------%s-----------------------\n\n", label)
	fmt.Println(bannerStartLine)
	//a, _ := json.Marshal(tw.FileTree)
	a, _ := json.MarshalIndent(tw.FileTree, "", "  ")
	fmt.Println(string(a))
	fmt.Println(bannerEndLine)
}

func (tw *TreeWatcher) Close() {
	err := tw.Watcher.Close()
	if err != nil {
		log.Error(err)
	}
}

func (tw *TreeWatcher) Remove(path connector.Path) (*filenode.FileNode, error) {
	eventPath := path.ExcludePath(tw.ParentPath)
	node, err := tw.FileTree.Remove(eventPath)
	if err != nil {
		return nil, err
	}

	err = tw.removeWatcherPath(path.String())
	if err != nil {
		return nil, errors.New("path not deleted in watchlist")
	}

	return node, err
}

func (tw *TreeWatcher) Move(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error) {
	node, err := tw.FileTree.Move(fromPath.ExcludePath(tw.ParentPath), toPath.ExcludePath(tw.ParentPath))
	if err != nil {
		return nil, err
	}
	return node, err
}

func (tw *TreeWatcher) Write(path connector.Path) (*filenode.FileNode, error) {
	if !path.IsDir() {
		eventPath := path.ExcludePath(tw.ParentPath)
		node, err := tw.FileTree.Update(eventPath, path)
		return node, err
	}
	return nil, nil
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
							tw.Errors <- err
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
	var err error
	node, err := tw.FileTree.Rename(fromPath.ExcludePath(tw.ParentPath), toPath.ExcludePath(tw.ParentPath))
	if err != nil {
		return nil, err
	}

	if toPath.IsDir() {
		err = tw.removeWatcherPath(fromPath.String())
		if err != nil {
			return nil, errors.New("path not deleted in watchlist")
		}
		err = tw.Watcher.Add(toPath.String())
		if err != nil {
			return nil, err
		}
	}
	return node, err
}

func (tw *TreeWatcher) Handler(e event.Event, extras ...*filenode.ExtraPayload) (*EventTransaction, error) {
	tw.Lock()
	defer tw.Unlock()

	var err error
	var node *filenode.FileNode
	var extra *filenode.ExtraPayload

	if len(extras) > 0 {
		extra = extras[0]
	}

	switch e.Type {
	case event.Remove:
		node, err = tw.Remove(e.FromPath)
		break
	case event.Write:
		node, err = tw.Write(e.FromPath)
		break
	case event.Create:
		node, err = tw.Create(e.FromPath, extra)
		break
	case event.Rename:
		node, err = tw.Rename(e.FromPath, e.ToPath)
		break
	case event.Move:
		node, err = tw.Move(e.FromPath, e.ToPath)
		break
	default:
		errorMsg := fmt.Sprintf("unhandled event: op:%s, path:%s", e.Type, e.FromPath)
		err = errors.New(errorMsg)
		break
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
			tw.Errors <- err
			if !ok {
				return
			}
		}
	}
}

func (tw *TreeWatcher) Start() {
	tw.IgniterReloadCtx, tw.IgniterReloadFunc = context.WithCancel(context.Background())
	// EventManager's working range
	ticker := time.NewTicker(2 * time.Second)

	go tw.start(ticker)
	go tw.Watch()
}

func (tw *TreeWatcher) start(ticker *time.Ticker) {
	for {
		select {
		case <-tw.IgniterReloadCtx.Done():
			tw.IgniterReloadCtx, tw.IgniterReloadFunc = context.WithCancel(context.Background())
			tw.Start()
			return
		case _ = <-ticker.C:
			if tw.EventManager.StackLength() > 0 {
				newEvents := tw.EventManager.Process()
				for _, e := range newEvents {
					txn, err := tw.Handler(e)
					if err != nil {
						tw.Errors <- err
					}
					tw.Events <- *txn

				}
			}
		}
	}
}

func (tw *TreeWatcher) Stop() {
	err := tw.Watcher.Close()
	if err != nil {
		log.Error(err)
	}
	close(tw.Events)
	close(tw.Errors)
}

func (tw *TreeWatcher) removeWatcherPath(fsPath string) error {
	/*
		When you see this function, why use watcher.Remove?
		You may think we need this function instead;
		The reason is; When we want to remove a path in watcher, we get an error like this.
		we encounter `GetFileAttributes:The system cannot find the file specified`
		As a solution, we generate the watchlist of the watcher again. This way is not good,
		but the situation needs to be saved.
	*/
	var err error
	var watcher *fsnotify.Watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	currentPathList := tw.Watcher.WatchList()
	for i := 0; i < len(currentPathList); i++ {
		if currentPathList[i] != fsPath {
			_ = watcher.Add(currentPathList[i])
		}
	}

	_ = tw.Watcher.Close()
	tw.Watcher = watcher
	tw.IgniterReloadFunc()
	return nil
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
		Subs: []*filenode.FileNode{},
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
		Events:       make(chan EventTransaction, 10),
		Errors:       make(chan error, 10),
	}

	e := event.Event{FromPath: path, Type: event.Create}
	txn, err := tw.Handler(e)
	if err != nil {
		tw.Errors <- err
		return nil, nil, err
	}
	tw.Start()
	return &tw, txn, nil
}
