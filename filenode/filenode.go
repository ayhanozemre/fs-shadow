package filenode

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/connector"
	"github.com/ayhanozemre/fs-shadow/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

func (fn *FileNode) Remove(absolutePath connector.Path) error {
	fileName := absolutePath.Name()
	node := fn.Search(absolutePath.ParentPath().String())
	if node == nil {
		return errors.New("FileNode not found")
	}
	if len(node.Subs) == 0 {
		return errors.New("Subs nodes not found")
	}
	for nodeIndex, sub := range node.Subs {
		if sub.Name == fileName {
			node.Subs = append(node.Subs[:nodeIndex], node.Subs[nodeIndex+1:]...)
			return nil
		}
	}
	return nil
}

func (fn *FileNode) Update(treePath connector.Path, absolutePath connector.Path) error {
	node := fn.Search(treePath.String())
	if node == nil {
		return errors.New("FileNode not found")
	}
	err := node.SumUpdate(absolutePath)
	if err != nil {
		return err
	}
	return nil
}

func (fn *FileNode) Create(path connector.Path, absolutePath connector.Path, ch chan connector.Path) error {
	sum, err := utils.Sum(absolutePath)
	if err != nil {
		return err
	}

	parentNode := fn.Search(path.ParentPath().String())
	if parentNode == nil {
		var wg sync.WaitGroup
		WalkOnFsPath(fn, absolutePath, &wg, ch)
		wg.Wait()
		return nil
	}

	absolutePathInfo := absolutePath.Info()
	meta := MetaData{
		IsDir:      absolutePath.IsDir(),
		Sum:        sum,
		Size:       absolutePathInfo.Size,
		CreatedAt:  absolutePathInfo.CreatedAt,
		Permission: absolutePathInfo.Permission,
	}
	node := FileNode{Name: path.Name(), Meta: meta}
	parentNode.Subs = append(parentNode.Subs, &node)
	if absolutePath.IsDir() {
		var wg sync.WaitGroup
		WalkOnFsPath(&node, absolutePath, &wg, ch)
		wg.Wait()
	}
	return nil
}

func (fn *FileNode) SumUpdate(absolutePath connector.Path) error {
	sum, err := utils.Sum(absolutePath)
	if err != nil {
		return err
	}
	fn.Meta.Sum = sum
	return nil
}

func (fn *FileNode) Search(path string) *FileNode {
	pathExp := strings.Split(path, "/")
	if fn.Name == pathExp[0] && len(pathExp) == 1 {
		return fn
	}
	if fn.Name == pathExp[0] {
		if len(pathExp) != 1 {
			newPath := filepath.Join(pathExp[1:]...)
			var wg sync.WaitGroup
			var wantedNode *FileNode
			for _, sub := range fn.Subs {
				wg.Add(1)
				go func(sub *FileNode) {
					node := sub.Search(newPath)
					if node != nil {
						wantedNode = node
					}
					wg.Done()
				}(sub)
			}
			wg.Wait()
			if wantedNode != nil {
				return wantedNode
			}
		}
	}
	return nil
}

func (fn *FileNode) JsonImport(tree string) error {
	/*
		{"file", "op"}

	*/
	return nil
}

func (fn *FileNode) JsonExport() (string, error) {
	dump, err := json.Marshal(fn)
	if err != nil {
		return "", err
	}
	return string(dump), nil
}

func (fn *FileNode) JsonUpdate() error {

	return nil
}

func WalkOnFsPath(root *FileNode, absolutePath connector.Path, wg *sync.WaitGroup, ch chan connector.Path) {
	ch <- absolutePath
	files, _ := ioutil.ReadDir(absolutePath.String())
	for _, path := range files {
		newAbsolutePath := connector.NewFSPath(filepath.Join(absolutePath.String(), path.Name()))
		mode := fmt.Sprintf("%d", path.Mode())

		sum, err := utils.Sum(newAbsolutePath)
		if err != nil {
			log.Error("sum error:", newAbsolutePath)
		}

		newNode := FileNode{
			Name: path.Name(),
			Meta: MetaData{
				IsDir:      path.IsDir(),
				Sum:        sum,
				Size:       path.Size(),
				CreatedAt:  path.ModTime().Unix(),
				Permission: mode,
			},
		}
		root.Subs = append(root.Subs, &newNode)

		if path.IsDir() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				WalkOnFsPath(&newNode, newAbsolutePath, wg, ch)
				return
			}()
		}
	}
}
