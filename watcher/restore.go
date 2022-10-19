package watcher

import (
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
)

func CreateFileNodeWithTransactions(tbl [][]byte) (*filenode.FileNode, error) {
	var err error
	var root *filenode.FileNode
	uuidTable := make(map[string]*filenode.FileNode)
	for i := 0; i < len(tbl); i++ {
		txn := EventTransaction{}
		err = txn.Decode(tbl[i])
		if err != nil {
			return nil, err
		}

		node := txn.toFileNode()
		if node.ParentUUID == "" {
			root = node
		}

		switch txn.Type {
		case event.Create:
			uuidTable[node.UUID] = node
			if parent, ok := uuidTable[node.ParentUUID]; ok {
				parent.Subs = append(parent.Subs, node)
			}
		case event.Rename:
			currentNode := uuidTable[node.UUID]
			currentNode.Name = node.Name
			currentNode.Meta = node.Meta
		case event.Move:
			currentNode := uuidTable[node.UUID]
			_, _ = root.RemoveByUUID(currentNode.UUID, currentNode.ParentUUID)
			if parent, ok := uuidTable[node.ParentUUID]; ok {
				currentNode.ParentUUID = node.ParentUUID
				parent.Subs = append(parent.Subs, currentNode)
			}
		case event.Remove:
			delete(uuidTable, node.UUID)
			_, _ = root.RemoveByUUID(node.UUID, node.ParentUUID)
		}
	}
	return root, nil
}

func RestoreWatcherWithTransactions(tbl [][]byte, tw Watcher) error {
	tree, err := CreateFileNodeWithTransactions(tbl)
	if err != nil {
		return err
	}
	tw.Restore(tree)
	return nil
}
