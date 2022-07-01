package connector

import (
	"fmt"
	"os"
	"strings"
)

type FSPath struct {
	p string
}

func NewFSPath(p string) *FSPath {
	return &FSPath{p: p}
}

func (path FSPath) IsVirtual() bool {
	return false
}

func (path FSPath) IsDir() bool {
	fInfo, err := os.Stat(path.String())
	if err != nil {
		return false
	}
	return fInfo.IsDir()
}

func (path *FSPath) Info() *FileInfo {
	p, _ := os.Stat(path.String())
	return &FileInfo{
		IsDir:      p.IsDir(),
		Size:       p.Size(),
		CreatedAt:  p.ModTime().Unix(),
		Permission: fmt.Sprintf("%d", p.Mode()),
	}
}

func (path *FSPath) Exists() bool {
	if _, err := os.Stat(path.String()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (path *FSPath) Name() string {
	parts := strings.Split(path.String(), "/")
	lastName := parts[len(parts)-1]
	return lastName
}

func (path *FSPath) String() string {
	return string(path.p)
}

func (path *FSPath) ParentPath() Path {
	parts := strings.Split(path.String(), "/")
	absolutePath := strings.Join(parts[:len(parts)-1], "/")
	return NewFSPath(absolutePath)
}

func (path *FSPath) ExcludePath(excPath Path) Path {
	eventAbsolutePath := strings.ReplaceAll(path.String(), excPath.String(), "")
	eventAbsolutePath = strings.Trim(eventAbsolutePath, "/")
	return NewFSPath(eventAbsolutePath)
}
