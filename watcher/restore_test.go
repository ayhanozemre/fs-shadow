package watcher

import (
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func generateTransactionBytes() [][]byte {
	// generate events
	ets := []EventTransaction{
		{Name: "test-1", UUID: "r1", Type: event.Create},
		{Name: "s-test-1", UUID: "s1", ParentUUID: "r1", Type: event.Create},
		{Name: "ss-test-1", UUID: "ss1", ParentUUID: "s1", Type: event.Create},
		{Name: "s-test-2", UUID: "s2", ParentUUID: "r1", Type: event.Create},
		{Name: "s-test-2-rename", UUID: "s2", ParentUUID: "r1", Type: event.Rename},
		{Name: "s-test-2-rename", UUID: "s2", ParentUUID: "r1", Type: event.Move},
	}

	// events to byte arrays
	var tbl [][]byte
	for i := 0; i < len(ets); i++ {
		b, _ := ets[i].Encode()
		tbl = append(tbl, b)
	}
	return tbl
}

func Test_CreateFileNodeWithTransaction(t *testing.T) {
	tbl := generateTransactionBytes()
	// process
	tree, err := CreateFileNodeWithTransactions(tbl)
	assert.Equal(t, nil, err, "tree creation error")
	assert.Equal(t, "test-1", tree.Name, "tree name is wrong")
	assert.Equal(t, "s-test-1", tree.Subs[0].Name, "first sub node name is wrong")
}

func Test_RestoreWatcherWithTransactions(t *testing.T) {
	root := "fs-shadow"
	tw, _, err := NewVirtualPathWatcher(root, &filenode.ExtraPayload{UUID: uuid.NewString()})
	assert.Equal(t, nil, err, "watcher creation error")

	tbl := generateTransactionBytes()
	err = RestoreWatcherWithTransactions(tbl, tw)
	assert.Equal(t, nil, err, "restoration error")
	assert.Equal(t, "test-1", tw.FileTree.Name, "tree name is wrong")
	assert.Equal(t, "s-test-1", tw.FileTree.Subs[0].Name, "first sub node name is wrong")
}
