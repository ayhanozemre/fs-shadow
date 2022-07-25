package filenode

import (
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/path"
	"github.com/ayhanozemre/fs-shadow/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

type FileNode struct {
	Subs       []*FileNode `json:"subs"`
	Name       string      `json:"name"`
	UUID       string      `json:"uuid"`
	ParentUUID string      `json:"parent_uuid"`
	Meta       MetaData    `json:"-"`
}

func (fn *FileNode) Rename(fromPath connector.Path, toPath connector.Path) (*FileNode, error) {
	node := fn.Search(fromPath.String())
	if node == nil {
		return node, errors.New("FileNode not found")
	}
	node.Name = toPath.Name()
	return node, nil
}

func (fn *FileNode) Remove(fromPath connector.Path) (deletedNode *FileNode, err error) {
	fileName := fromPath.Name()
	parentNode := fn.Search(fromPath.ParentPath().String())
	return fn._remove(parentNode, fileName)
}

// RemoveByUUID sdlkfjsdlkjf
// sdfklsjdlkfjs
// sdlkfslkdjf
func (fn *FileNode) RemoveByUUID(uuid string, parentUUID string) (*FileNode, error) {
	parentNode := fn.SearchByUUID(parentUUID)
	return fn._remove(parentNode, uuid, "uuid")
}

func (fn *FileNode) _remove(parentNode *FileNode, uniq string, searchField ...string) (deletedNode *FileNode, err error) {
	if parentNode == nil {
		return nil, errors.New("FileNode not found")
	}
	if len(parentNode.Subs) == 0 {
		return nil, errors.New("Subs nodes not found")
	}
	for nodeIndex, sub := range parentNode.Subs {
		var lookupValue string
		if len(searchField) > 0 && searchField[0] == "uuid" {
			lookupValue = sub.UUID
		} else {
			lookupValue = sub.Name
		}
		if lookupValue == uniq {
			deletedNode = parentNode.Subs[nodeIndex]
			parentNode.Subs = append(parentNode.Subs[:nodeIndex], parentNode.Subs[nodeIndex+1:]...)
			return deletedNode, nil
		}
	}
	return nil, err
}

func (fn *FileNode) Update(fromPath connector.Path, absolutePath connector.Path) (*FileNode, error) {
	node := fn.Search(fromPath.String())
	if node == nil {
		return nil, errors.New("FileNode not found")
	}
	err := node.SumUpdate(absolutePath)
	if err != nil {
		return node, err
	}
	return node, nil
}

func (fn *FileNode) UpdateWithExtra(extra ExtraPayload) {
	fn.UUID = extra.UUID
}

func (fn *FileNode) Create(fromPath connector.Path, absolutePath connector.Path, ch ...chan connector.Path) (*FileNode, error) {
	var sum string
	var err error

	if !absolutePath.IsVirtual() {
		sum, err = utils.Sum(absolutePath)
		if err != nil {
			return nil, err
		}
	}

	parentNode := fn.Search(fromPath.ParentPath().String())
	if parentNode == nil {
		if !fromPath.IsVirtual() {
			var wg sync.WaitGroup
			WalkOnFsPath(fn, absolutePath, &wg, ch[0])
			wg.Wait()
		}
		return fn, nil
	}

	var _uuid string
	if !fromPath.IsVirtual() {
		_uuid = uuid.NewString()
	}
	absolutePathInfo := absolutePath.Info()
	meta := MetaData{
		IsDir:      absolutePath.IsDir(),
		Sum:        sum,
		Size:       absolutePathInfo.Size,
		CreatedAt:  absolutePathInfo.CreatedAt,
		Permission: absolutePathInfo.Permission,
	}
	node := FileNode{
		Name:       fromPath.Name(),
		UUID:       _uuid,
		ParentUUID: parentNode.UUID,
		Meta:       meta,
	}
	parentNode.Subs = append(parentNode.Subs, &node)
	if absolutePath.IsDir() && !absolutePath.IsVirtual() {
		var wg sync.WaitGroup
		WalkOnFsPath(&node, absolutePath, &wg, ch[0])
		wg.Wait()
	}
	return &node, nil
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
	pathExp := strings.Split(path, connector.Separator)
	if fn.Name == pathExp[0] && len(pathExp) == 1 {
		return fn
	}
	if fn.Name == pathExp[0] && len(pathExp) != 1 {
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

	return nil
}

func (fn *FileNode) SearchByUUID(uuid string) *FileNode {
	if fn.UUID == uuid {
		return fn
	}
	if len(fn.Subs) == 0 {
		return nil
	}
	var wantedNode *FileNode
	var wg sync.WaitGroup
	for _, sn := range fn.Subs {
		wg.Add(1)
		go func(sn *FileNode) {
			node := sn.SearchByUUID(uuid)
			if node != nil {
				wantedNode = node
			}
			wg.Done()
		}(sn)
	}
	wg.Wait()
	return wantedNode
}

func WalkOnFsPath(root *FileNode, absolutePath connector.Path, wg *sync.WaitGroup, ch chan connector.Path) {
	ch <- absolutePath
	files, _ := ioutil.ReadDir(absolutePath.String())
	for _, path := range files {
		newAbsolutePath := connector.NewFSPath(filepath.Join(absolutePath.String(), path.Name()))
		mode := fmt.Sprintf("%d", path.Mode())

		sum, err := utils.Sum(newAbsolutePath)
		if err != nil {
			log.Error("sum error:", newAbsolutePath, err)
		}

		newNode := FileNode{
			Name: path.Name(),
			UUID: uuid.NewString(),
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
				WalkOnFsPath(&newNode, newAbsolutePath, wg, ch)
				wg.Done()
			}()
		}
	}
}
