package connector

import (
	"strings"
)

type VirtualPath struct {
	p     string
	isDir bool
}

func NewVirtualPath(p string, isDir bool) *VirtualPath {
	return &VirtualPath{p: p, isDir: isDir}
}

func (path VirtualPath) IsVirtual() bool {
	return true
}

func (path VirtualPath) IsDir() bool {
	return path.isDir
}

func (path VirtualPath) Info() *FileInfo {
	return &FileInfo{}
}

func (path *VirtualPath) Exists() bool {
	return true
}

func (path *VirtualPath) Name() string {
	parts := strings.Split(path.String(), Separator)
	lastName := parts[len(parts)-1]
	return lastName
}

func (path *VirtualPath) String() string {
	return string(path.p)
}

func (path *VirtualPath) ParentPath() Path {
	parts := strings.Split(path.String(), Separator)
	absolutePath := strings.Join(parts[:len(parts)-1], Separator)
	return NewVirtualPath(absolutePath, true)
}

func (path *VirtualPath) ExcludePath(excPath Path) Path {
	eventAbsolutePath := strings.ReplaceAll(path.String(), excPath.String(), "")
	eventAbsolutePath = strings.Trim(eventAbsolutePath, Separator)
	return NewVirtualPath(eventAbsolutePath, path.IsDir())
}
