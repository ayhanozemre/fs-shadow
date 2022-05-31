package main

import (
	"encoding/json"
	"errors"
	"fmt"
	filenode "github.com/ayhanozemre/fs-shadow/internal/filenode"
	"github.com/ayhanozemre/fs-shadow/utils"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Event struct {
	Path  string
	Op    string
	Error error
}

type TreeWatcher struct {
	FileTree   *filenode.FileNode
	Watcher    *fsnotify.Watcher
	Path       string
	ParentPath string

	Events chan Event // bu channel'i servislere verecegiz. not implemented
	sync.Mutex
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

func (tw *TreeWatcher) Remove(path utils.Path) error {
	eventPath := path.ExcludePath(tw.ParentPath)
	err := tw.FileTree.Remove(eventPath)
	if err == nil && path.IsDir() {
		err := tw.Watcher.Remove(path.String())
		if err != nil {
			return err
		}
	}
	return err
}

func (tw *TreeWatcher) Write(path utils.Path) error {
	if !path.IsDir() {
		eventPath := path.ExcludePath(tw.ParentPath)
		err := tw.FileTree.Update(eventPath, path.String())
		return err
	}
	return nil
}

func (tw *TreeWatcher) Create(path utils.Path) error {
	if !path.Exists() {
		return errors.New("file path does not exist")
	}

	eventPath := path.ExcludePath(tw.ParentPath)
	eventCh := make(chan string)

	go func() {
		for {
			select {
			case p := <-eventCh:
				if p != "" {
					newFile := utils.Path(p)
					if newFile.IsDir() {
						err := tw.Watcher.Add(p)
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

	err := tw.FileTree.Create(eventPath, path.String(), eventCh)
	eventCh <- ""
	close(eventCh)
	return err
}

func (tw *TreeWatcher) Rename(path utils.Path) error {
	if !path.Exists() {
		return tw.Remove(path)
	}
	return nil
}

func (tw *TreeWatcher) EventHandler(op fsnotify.Op, path string) (err error) {
	tw.Lock()
	defer tw.Unlock()

	if op == fsnotify.Chmod {
		return nil
	}
	pathIns := utils.Path(path)

	switch op {
	case fsnotify.Remove:
		err = tw.Remove(pathIns)
	case fsnotify.Write:
		err = tw.Write(pathIns)
	case fsnotify.Create:
		err = tw.Create(pathIns)
	case fsnotify.Rename:
		err = tw.Rename(pathIns)
	default:
		return errors.New("unhandled event")
	}
	return nil
}

func (tw *TreeWatcher) Watch() {
	for {
		select {
		case event, ok := <-tw.Watcher.Events:
			if !ok {
				return
			}
			err := tw.EventHandler(event.Op, event.Name)
			if err != nil {
				fmt.Println(err)
			} else {
				tw.PrintTree("EVENT MANAGER")
			}
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
	go tw.Watch()
}

func NewPathWatcher(path utils.Path) (*TreeWatcher, error) {
	if !path.IsDir() {
		err := errors.New("input path is not directory")
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
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
		FileTree:   &root,
		ParentPath: path.ParentPath(),
		Path:       path.String(),
		Watcher:    watcher,
	}
	err = tw.Create(path)
	if err != nil {
		return nil, err
	}
	tw.Start()
	return &tw, nil
}

func main() {
	input := utils.Path("/home/wade/Desktop/TransferChain")
	tw, err := NewPathWatcher(input)
	fmt.Println("err->", err)
	tw.PrintTree("INIT TREE")
	done := make(chan bool)
	<-done
	tw.Close()
}
