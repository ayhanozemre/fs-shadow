package filenode

import (
	"github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
)

type MetaData struct {
	IsDir      bool   `json:"is_dir"`
	Sum        string `json:"sum"`
	Size       int64  `json:"size"`
	CreatedAt  int64  `json:"createdAt"`
	Permission string `json:"permission"`
}

type FileNode struct {
	Subs       []*FileNode `json:"subs"`
	Name       string      `json:"name"`
	UUID       string      `json:"uuid"`
	ParentUUID string      `json:"parent_uuid"`
	Meta       MetaData    `json:"-"`
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
