package watcher

import (
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/vmihailenco/msgpack/v5"
)

type Watcher interface {
	PrintTree(label string) // for debug
	Start()
	Watch()
	Close()
	GetEvents() <-chan EventTransaction
	GetErrors() <-chan error
	Handler(event event.Event, extra ...*filenode.ExtraPayload) (*EventTransaction, error)
	Create(fromPath connector.Path, extra *filenode.ExtraPayload) (*filenode.FileNode, error)
	Write(fromPath connector.Path) (*filenode.FileNode, error)
	Rename(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error)
	Move(fromPath connector.Path, toPath connector.Path) (*filenode.FileNode, error)
	Remove(fromPath connector.Path) (*filenode.FileNode, error)
	Restore(tree *filenode.FileNode)
}

type EventTransaction struct {
	Name       string
	Type       event.Type
	UUID       string
	ParentUUID string
	Meta       filenode.MetaData
}

func (t EventTransaction) Encode() ([]byte, error) {
	b, err := msgpack.Marshal(t)
	if err != nil {
		return nil, err
	}
	return b, err
}

func (t *EventTransaction) Decode(b []byte) error {
	err := msgpack.Unmarshal(b, t)
	if err != nil {
		return err
	}
	return err
}

func (t *EventTransaction) toFileNode() *filenode.FileNode {
	return &filenode.FileNode{
		Name:       t.Name,
		UUID:       t.UUID,
		ParentUUID: t.ParentUUID,
		Meta:       t.Meta,
	}
}

func makeEventTransaction(node filenode.FileNode, event event.Type) *EventTransaction {
	return &EventTransaction{
		Type:       event,
		Name:       node.Name,
		Meta:       node.Meta,
		UUID:       node.UUID,
		ParentUUID: node.ParentUUID,
	}
}

func NewFSWatcher(fsPath string) (Watcher, *EventTransaction, error) {
	return NewPathWatcher(fsPath)
}

func NewVirtualWatcher(fsPath string, extra *filenode.ExtraPayload) (Watcher, *EventTransaction, error) {
	return NewVirtualPathWatcher(fsPath, extra)
}
