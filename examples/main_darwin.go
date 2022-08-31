package main

import (
	"fmt"
	"github.com/ayhanozemre/fs-shadow/watcher"
	log "github.com/sirupsen/logrus"
)

func main() {
	// not completed
	log.SetLevel(log.DebugLevel)
	tw, _, err := watcher.NewFSWatcher("/tmp/fs-shadow")

	if err == nil {
		go func() {
			ch := tw.GetEvents()
			err := tw.GetErrors()
			for {
				select {
				case p := <-ch:
					log.Debug(fmt.Sprintf("Event-> Name:%s UUID:%s", p.Name, p.UUID))
				case e := <-err:
					log.Debug("Error->", e)
				}
				tw.PrintTree("TREE")
			}
		}()
		done := make(chan bool)
		<-done
	} else {
		log.Panic(err)
	}
	tw.Stop()
}
