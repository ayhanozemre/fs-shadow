package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ayhanozemre/fs-shadow/event"
	"github.com/ayhanozemre/fs-shadow/filenode"
	connector "github.com/ayhanozemre/fs-shadow/path"
	"github.com/ayhanozemre/fs-shadow/watcher"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func textToEventType(text string) (event.Type, error) {
	switch strings.ToLower(text) {
	case "create":
		return event.Create, nil
	case "remove":
		return event.Remove, nil
	case "rename":
		return event.Rename, nil
	case "move":
		return event.Move, nil
	}
	return "", errors.New("invalid event type")
}

func main() {
	var err error
	var tw watcher.Watcher
	log.SetLevel(log.DebugLevel)
	extra := &filenode.ExtraPayload{UUID: uuid.NewString()}
	tw, _, err = watcher.NewVirtualWatcher("root", extra)

	if err == nil {
		tw.PrintTree("INIT TREE")
		fmt.Println("First Node name is 'root'")
		fmt.Println("Examples:\n create /root/sub\n rename /root/sub /root/test\n move /root/sub /root/test\n remove /root/test")
		fmt.Println("Press Q/q to quit the loop")

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("~$ ")
			scanner.Scan()
			text := scanner.Text()
			if len(text) != 0 && strings.ToLower(text) != "q" {
				exp := strings.Split(text, " ")
				expSize := len(exp)

				if expSize < 2 {
					log.Debug("invalid argument")
					continue
				}
				var eventType event.Type
				eventType, err = textToEventType(exp[0])
				if err != nil {
					log.Debug("Error:", err)
				}

				fromPath := connector.NewVirtualPath(exp[1], true)
				var toPath connector.Path
				if expSize == 3 {
					toPath = connector.NewVirtualPath(exp[2], true)
				}

				e := event.Event{FromPath: fromPath, ToPath: toPath, Type: eventType}
				_, err = tw.Handler(e, &filenode.ExtraPayload{UUID: uuid.NewString()})
				if err != nil {
					log.Debug("Error:", err)
					continue
				}
				tw.PrintTree("TREE")
			} else {
				break
			}

		}
	} else {
		log.Panic(err)
	}

}
