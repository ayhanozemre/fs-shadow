package watcher

import (
	"github.com/ayhanozemre/fs-shadow/connector"
	"github.com/ayhanozemre/fs-shadow/filenode"
	"github.com/fsnotify/fsnotify"
)

func NewVirtualPathWatcher(fs_path string) (*TreeWatcher, error) {
	var err error
	path := connector.NewVirtualPath(fs_path)

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
		Path:       path,
		Watcher:    watcher,
	}
	err = tw.Create(path)
	if err != nil {
		return nil, err
	}
	tw.Start()
	return &tw, nil
}
