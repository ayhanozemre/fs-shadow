package event

import (
	"fmt"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
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
	log.SetLevel(log.DebugLevel)

	handler := newEventHandler()
	testFolder := "fs-shadow-test"

	file := filepath.Join(testFolder, "test.txt")
	newFolder := filepath.Join(testFolder, "New Folder")
	folder := filepath.Join(testFolder, "test")

	//----------------------------------------------------
	// create process
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	checkSingleEventResult(t, "[1] create", Event{FromPath: connector.NewFSPath(folder), Type: Create}, handler.Process())

	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Write}, "")
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Write}, "")
	checkSingleEventResult(t, "[2] create", Event{FromPath: connector.NewFSPath(file), Type: Create}, handler.Process())

	handler.Append(fsnotify.Event{Name: file, Op: fsnotify.Create}, "")
	checkSingleEventResult(t, "[3] create", Event{FromPath: connector.NewFSPath(file), Type: Create}, handler.Process())

	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	checkSingleEventResult(t, "[4] create", Event{FromPath: connector.NewFSPath(folder), Type: Create}, handler.Process())

	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	checkSingleEventResult(t, "[5] create", Event{FromPath: connector.NewFSPath(folder), Type: Create}, handler.Process())

	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Rename}, "")
	checkSingleEventResult(t, "[5] create", Event{FromPath: connector.NewFSPath(folder), Type: Create}, handler.Process())

	//----------------------------------------------------
	// remove process
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Remove}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Remove}, "")
	checkSingleEventResult(t, "[1] remove", Event{FromPath: connector.NewFSPath(folder), Type: Remove}, handler.Process())

	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Remove}, "")
	checkSingleEventResult(t, "[2] create", Event{FromPath: connector.NewFSPath(folder), Type: Remove}, handler.Process())

	//----------------------------------------------------
	// rename
	// case 1
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Write}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	checkSingleEventResult(t, "[1] rename", Event{
		FromPath: connector.NewFSPath(newFolder),
		ToPath:   connector.NewFSPath(folder),
		Type:     Rename}, handler.Process())

	// case 2
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	checkSingleEventResult(t, "[2] rename", Event{
		FromPath: connector.NewFSPath(newFolder),
		ToPath:   connector.NewFSPath(folder),
		Type:     Rename}, handler.Process())

	// case 3
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Write}, "")
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Write}, "")
	checkSingleEventResult(t, "[3] rename", Event{
		FromPath: connector.NewFSPath(newFolder),
		ToPath:   connector.NewFSPath(folder),
		Type:     Rename}, handler.Process())

	// case 4
	handler.Append(fsnotify.Event{Name: newFolder, Op: fsnotify.Rename}, "")
	handler.Append(fsnotify.Event{Name: folder, Op: fsnotify.Create}, "")
	checkSingleEventResult(t, "[4] rename", Event{
		FromPath: connector.NewFSPath(newFolder),
		ToPath:   connector.NewFSPath(folder),
		Type:     Rename}, handler.Process())

}

func Test_EventQueue(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	handler := newEventHandler()
	testFolder := "fs-shadow-test"
	folder := connector.NewFSPath(filepath.Join(testFolder, "test"))
	folder1 := connector.NewFSPath(filepath.Join(testFolder, "test1"))

	// mkdir fs-shadow-test/test
	handler.Append(fsnotify.Event{Name: folder.String(), Op: fsnotify.Create}, "1")

	// rm /tmp/fs-shadow/test1
	handler.Append(fsnotify.Event{Name: folder1.String(), Op: fsnotify.Remove}, "2")

	// mv fs-shadow-test/test .
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
		{FromPath: folder1, Type: Remove},
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
