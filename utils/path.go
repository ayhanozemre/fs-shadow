package utils

import (
	"os"
	"strings"
)

type Path string

func (path Path) IsDir() bool {
	fInfo, ok := os.Stat(path.String())
	if ok != nil {
		return false
	}
	return fInfo.IsDir()
}

func (path *Path) Exists() bool {
	if _, err := os.Stat(path.String()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (path *Path) Name() string {
	parts := strings.Split(path.String(), "/")
	lastName := parts[len(parts)-1]
	return lastName
}

func (path *Path) String() string {
	return string(*path)
}

func (path *Path) ParentPath() string {
	parts := strings.Split(path.String(), "/")
	absolutePath := strings.Join(parts[:len(parts)-1], "/")
	return absolutePath
}

func (path *Path) ExcludePath(excPath string) string {
	eventAbsolutePath := strings.ReplaceAll(path.String(), excPath, "")
	eventAbsolutePath = strings.Trim(eventAbsolutePath, "/")
	return eventAbsolutePath
}
