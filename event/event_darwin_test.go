package event

import (
	"fmt"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"testing"
)

func checkSingleEventResult(t *testing.T, name string, expect Event, result []Event) {
	if len(result) != 0 {
		firstResult := result[0]
		if expect.String() != result[0].String() {
			message := fmt.Sprintf("[%s] exceptType:%s resultType:%s", name, expect.Type, firstResult.Type)
			t.Fatalf(message)
		}
	}
}

func Test_SingleEvents(t *testing.T) {
	handler := newEventHandler()
	testFolder := "/tmp/fs-shadow"

	file := filepath.Join(testFolder, "test.txt")
	file1 := filepath.Join(testFolder, "test1.txt")
	folder := filepath.Join(testFolder, "test")
	folder1 := filepath.Join(testFolder, "test1")

	// create test process folder
	_ = os.Mkdir(testFolder, os.ModePerm)
	//-------------------------------------------------------------------------------------
	// create process

	//mkdir /tmp/fs-shadow/test
	_ = os.Mkdir(folder, os.ModePerm)
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	checkSingleEventResult(t, "[1] create folder", Event{FromPath: connector.NewFSPath(folder), Type: Create}, handler.Process())
	_ = os.Remove(folder)

	//touch /tmp/fs-shadow/test.txt
	emptyFile, _ := os.Create(file)
	_ = emptyFile.Close()
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Chmod}, "")
	checkSingleEventResult(t, "[2] create file", Event{FromPath: connector.NewFSPath(file), Type: Create}, handler.Process())
	_ = os.Remove(file)

	/*
		// watcher inactive; mv /tmp/fs-shadow/test .
		//handler.stack = []fsnotify.Event{}
		_ = os.Mkdir(folder, os.ModePerm)
		handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
		handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
		checkSingleEventResult(t, "[3] create outside to inside", Event{FromPath: folder, Type: Create}, handler.Process())
		_ = os.Remove(folder)
		fmt.Println("size", len(handler.stack))

	*/

	// watcher active; mv /tmp/fs-shadow/test .
	_ = os.Mkdir(folder, os.ModePerm)
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Rename}, "")
	checkSingleEventResult(t, "[4] w create outside to inside", Event{FromPath: connector.NewFSPath(folder), Type: Create}, handler.Process())
	_ = os.Remove(folder)

	//-------------------------------------------------------------------------------------
	// remove process

	// watcher active; rm -rf
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Remove}, "")
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Remove}, "")
	checkSingleEventResult(t, "[1] w remove file", Event{FromPath: connector.NewFSPath(file), Type: Remove}, handler.Process())

	// watcher active; mv test /tmp/
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Rename}, "")
	checkSingleEventResult(t, "[2] w remove file", Event{FromPath: connector.NewFSPath(file), Type: Remove}, handler.Process())

	// watcher inactive. file or folder doesn't matter; mv test /tmp/
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Rename}, "")
	checkSingleEventResult(t, "[3] remove file", Event{FromPath: connector.NewFSPath(file), Type: Remove}, handler.Process())

	// watcher inactive; rm -rf
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Remove}, "")
	checkSingleEventResult(t, "[4] w remove file", Event{FromPath: connector.NewFSPath(file), Type: Remove}, handler.Process())

	//-------------------------------------------------------------------------------------
	// rename

	// watcher active; mv /tmp/test /tmp/fs-shadow/test1
	_ = os.Mkdir(folder1, os.ModePerm)
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder1, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Rename}, "")
	checkSingleEventResult(t, "[1] w rename folder", Event{FromPath: connector.NewFSPath(folder), ToPath: connector.NewFSPath(folder1), Type: Rename}, handler.Process())
	_ = os.Remove(folder1)

	// watcher inactive; mv /tmp/fs-shadow/test /tmp/fs-shadow/test1
	_ = os.Mkdir(folder1, os.ModePerm)
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder1, Op: fsnotify.Create}, "")
	checkSingleEventResult(t, "[2] rename folder", Event{FromPath: connector.NewFSPath(folder), ToPath: connector.NewFSPath(folder1), Type: Rename}, handler.Process())
	_ = os.Remove(folder1)

	// rename file; mv /tmp/test.txt /tmp/fs-shadow/test1.txt
	emptyFile, _ = os.Create(file1)
	_ = emptyFile.Close()
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: file1, Op: fsnotify.Create}, "")
	checkSingleEventResult(t, "[3] rename file", Event{FromPath: connector.NewFSPath(file), ToPath: connector.NewFSPath(file1), Type: Rename}, handler.Process())
	_ = os.Remove(file1)

	// remove test process folder
	_ = os.Remove(testFolder)
}

func Test_EventQueue(t *testing.T) {
	handler := newEventHandler()
	testFolder := "/tmp/fs-shadow"
	folder := connector.NewFSPath(filepath.Join(testFolder, "test"))

	// mkdir /tmp/fs-shadow/test
	handler.Append(fsnotify.Event{Name: folder.String(), Op: fsnotify.Create}, "1")

	// mv /tmp/fs-shadow/test /tmp/test
	//_ = os.Remove(folder)
	handler.Append(fsnotify.Event{Name: folder.String(), Op: fsnotify.Rename}, "2")

	// mv /tmp/fs-shadow/test .
	handler.Append(fsnotify.Event{Name: folder.String(), Op: fsnotify.Create}, "3")

	var results []Event
	for {
		r := handler.Process()
		if len(r) == 0 {
			break
		}
		results = append(results, r...)
	}
	expectResult := []Event{
		{FromPath: folder, Type: Create},
		{FromPath: folder, Type: Remove},
		{FromPath: folder, Type: Create},
	}
	if len(results) != len(expectResult) {
		t.Fatalf("Results(%d) and ExpectResult(%d) not equal", len(results), len(expectResult))
	}
	for i, expectValue := range expectResult {
		resultValue := results[i]
		if expectValue.String() != resultValue.String() {
			message := fmt.Sprintf("exceptType:%s resultType:%s", expectValue.Type, resultValue.Type)
			t.Fatalf(message)
		}
	}
}
