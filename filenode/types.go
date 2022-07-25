package filenode

import (
	"github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
)

type MetaData struct {
	IsDir      bool   `json:"is_dir"`
	Sum        string `json:"sum"`
	Size       int64  `json:"size"`
	CreatedAt  int64  `json:"created_at"`
	Permission string `json:"permission"`
}

type FileTree struct {
	Path    connector.Path
	Tree    *FileNode
	Watcher *fsnotify.Watcher
}

type ExtraPayload struct {
	UUID         string
	AbsolutePath string
}
