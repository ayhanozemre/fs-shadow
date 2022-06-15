package connector

import (
	"strings"
)

type VirtualPath struct {
	p string
}

func NewVirtualPath(p string) *VirtualPath {
	return &VirtualPath{p: p}
}

func (path VirtualPath) IsVirtual() bool {
	return true
}

func (path VirtualPath) IsDir() bool {
	return false
}

func (path VirtualPath) Info() *FileInfo {
	return nil
}

func (path *VirtualPath) Exists() bool {
	return false
}

func (path *VirtualPath) Name() string {
	parts := strings.Split(path.String(), "/")
	lastName := parts[len(parts)-1]
	return lastName
}

func (path *VirtualPath) String() string {
	return string(path.p)
}

func (path *VirtualPath) ParentPath() Path {
	parts := strings.Split(path.String(), "/")
	absolutePath := strings.Join(parts[:len(parts)-1], "/")
	return NewVirtualPath(absolutePath)
}

func (path *VirtualPath) ExcludePath(excPath Path) Path {
	eventAbsolutePath := strings.ReplaceAll(path.String(), excPath.String(), "")
	eventAbsolutePath = strings.Trim(eventAbsolutePath, "/")
	return NewVirtualPath(eventAbsolutePath)
}
