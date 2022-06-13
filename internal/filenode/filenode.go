package filenode

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

/*
	fileNode'n kullandigi path type kesinlikle fs'e bagli bir islem yapmamali.
	yeni bir Path type olusturulmali VPath gibi... burada fs'e depend olunmamali.
*/
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

// bu method fs'e depend, refactor.
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
		WalkOnFsPath(fn, absolutePath, &wg, ch)
		wg.Wait()
		return nil
	}

	meta := MetaData{
		IsDir: aPath.IsDir(),
		Sum:   sum,
		//Size:       aPath.Size(),
		//CreatedAt:  aPath.ModTime(),
		//Permission: aPath.Mode(),
	}
	node := FileNode{Name: eventPath.Name(), Meta: meta}
	parentNode.Subs = append(parentNode.Subs, &node)
	if aPath.IsDir() {
		var wg sync.WaitGroup
		WalkOnFsPath(&node, absolutePath, &wg, ch)
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

func WalkOnFsPath(root *FileNode, absolutePath string, wg *sync.WaitGroup, ch chan string) {
	ch <- absolutePath
	files, _ := ioutil.ReadDir(absolutePath)
	for _, path := range files {
		newAbsolutePath := filepath.Join(absolutePath, path.Name())
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
