package watcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/event"
	filenode "github.com/ayhanozemre/fs-shadow/filenode"
	"github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"strings"
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

func (tw *TreeWatcher) SearchByPath(path string) *filenode.FileNode {
	return tw.FileTree.Search(path)
}

func (tw *TreeWatcher) SearchByUUID(uuid string) *filenode.FileNode {
	return tw.FileTree.SearchByUUID(uuid)
}

func (tw *TreeWatcher) PrintTree(label string) {

	bannerStartLine := fmt.Sprintf("----------------%s----------------", label)
	bannerEndLine := fmt.Sprintf("----------------%s----------------\n\n", label)
	fmt.Println(bannerStartLine)
	a, _ := json.MarshalIndent(tw.FileTree, "", "  ")
	//a, _ := json.Marshal(tw.FileTree)
	fmt.Println(string(a))
	fmt.Println(bannerEndLine)
}

func (tw *TreeWatcher) Remove(path connector.Path) (*filenode.FileNode, error) {
	eventPath := path.ExcludePath(tw.ParentPath)
	node, err := tw.FileTree.Remove(eventPath)
	return node, err
}

func (tw *TreeWatcher) Write(path connector.Path) (*filenode.FileNode, error) {
	var node *filenode.FileNode
	var err error
	if !path.IsDir() {
		eventPath := path.ExcludePath(tw.ParentPath)
		node, err = tw.FileTree.Update(eventPath, path)
		return node, err
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
	node, err := tw.FileTree.Rename(fromPath.ExcludePath(tw.ParentPath), toPath.ExcludePath(tw.ParentPath))
	if err != nil {
		return nil, err
	}
	if toPath.IsDir() {
		err = tw.reloadWatcherForRename(fromPath.String(), toPath.String())
		if err != nil {
			return nil, err
		}
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

// Handler the 'extras' parameter is optional because we may need to move an external value to the node layer.
// sample; We want to parameterize the uuid from outside in VFS, but we don't want to do that in FS.
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
	log.Debug("start!")
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
						continue
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

func (tw *TreeWatcher) Restore(tree *filenode.FileNode) {
	tw.FileTree = tree
}

func isParentPath(a, b string) bool {
	aExp := strings.Split(a, "/") // parent possible?
	bExp := strings.Split(b, "/")
	if len(aExp) > len(bExp) {
		return false
	}
	for i := 0; i < len(aExp); i++ {
		if aExp[i] != bExp[i] {
			return false
		}
	}
	return true
}
func (tw *TreeWatcher) reloadWatcherForRename(fromPath string, toPath string) error {
	log.Debug("reload!")
	var err error
	var watcher *fsnotify.Watcher
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	currentPathList := tw.Watcher.WatchList()
	for i := 0; i < len(currentPathList); i++ {
		path := currentPathList[i]
		if isParentPath(fromPath, path) {
			//path != toPath && strings.HasPrefix(path, fromPath)
			path = strings.ReplaceAll(path, fromPath, toPath)
		}
		err = watcher.Add(path)
		if err != nil {
			return err
		}
	}

	err = tw.Watcher.Close()
	if err != nil {
		return err
	}
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
		UUID: uuid.NewString(),
		Meta: filenode.MetaData{
			IsDir: true,
		},
		Subs: []*filenode.FileNode{},
	}

	tw := TreeWatcher{
		FileTree:     &root,
		ParentPath:   path.ParentPath(),
		Path:         path,
		Watcher:      watcher,
		EventManager: event.NewEventHandler(),
		Events:       make(chan EventTransaction, 100),
		Errors:       make(chan error, 100),
	}
	e := event.Event{FromPath: path, Type: event.Create}
	txn, err := tw.Handler(e)
	if err != nil {
		tw.Errors <- err
		return nil, nil, err
	}
	tw.Start()
	tw.Events <- *txn
	return &tw, txn, nil
}
