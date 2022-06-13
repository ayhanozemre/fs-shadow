package filenode

import "github.com/fsnotify/fsnotify"

type MetaData struct {
	IsDir      bool   `json:"is_dir"`
	Sum        string `json:"sum"`
	Size       int64  `json:"size"`
	CreatedAt  int64  `json:"createdAt"`
	Permission string `json:"permission"`
}

type FileNode struct {
	Subs []*FileNode `json:"subs"`
	Name string      `json:"name"`
	Meta MetaData    `json:"-"`
}

type FileTree struct {
	Path    string
	Tree    *FileNode
	Watcher *fsnotify.Watcher
}
