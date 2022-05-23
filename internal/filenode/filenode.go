package filenode

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"walker/utils"
)

func (fn *FileNode) Remove(path string) error {
	tPath := utils.Path(path)
	fileName := tPath.Name()
	node := fn.Search(tPath.ParentPath())
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

func (fn *FileNode) Update(path string, absolutePath string) error {
	node := fn.Search(path)
	if node == nil {
		return errors.New("FileNode not found")
	}
	err := node.SumUpdate(absolutePath)
	if err != nil {
		return err
	}
	return nil
}

func (fn *FileNode) Create(path string, absolutePath string, ch chan string) error {
	sum, err := utils.Sum(absolutePath)
	if err != nil {
		return err
	}
	eventPath := utils.Path(path)
	aPath := utils.Path(absolutePath)
	parentNode := fn.Search(eventPath.ParentPath())
	if parentNode == nil {
		var wg sync.WaitGroup
		WalkOnPath(fn, absolutePath, &wg, ch)
		wg.Wait()
		return nil
	}

	meta := MetaData{
		IsDir: aPath.IsDir(),
		Sum:   sum,
	}
	node := FileNode{Name: eventPath.Name(), Meta: meta}
	parentNode.Subs = append(parentNode.Subs, &node)
	if aPath.IsDir() {
		var wg sync.WaitGroup
		WalkOnPath(&node, absolutePath, &wg, ch)
		wg.Wait()
	}
	return nil
}

func (fn *FileNode) SumUpdate(absolutePath string) error {
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
	newPath := filepath.Join(pathExp[1:]...)
	// goroutinelerle calismasi lazim
	if fn.Name == pathExp[0] {
		if len(pathExp) != 1 {
			for _, sub := range fn.Subs {
				node := sub.Search(newPath)
				if node != nil {
					return node
				}
			}
		}
	}
	return nil
}

func WalkOnPath(root *FileNode, absolutePath string, wg *sync.WaitGroup, ch chan string) {
	ch <- absolutePath
	files, _ := ioutil.ReadDir(absolutePath)
	for _, path := range files {
		newAbsolutePath := filepath.Join(absolutePath, path.Name())
		sum, err := utils.Sum(newAbsolutePath)
		if err != nil {
			log.Error("sum error:", newAbsolutePath)
		}
		newNode := FileNode{
			Name: path.Name(),
			Meta: MetaData{
				IsDir: path.IsDir(),
				Sum:   sum,
			},
		}
		root.Subs = append(root.Subs, &newNode)

		if path.IsDir() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				WalkOnPath(&newNode, newAbsolutePath, wg, ch)
				return
			}()
		}
	}
}
