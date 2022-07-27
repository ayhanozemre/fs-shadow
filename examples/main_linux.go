package main

import (
	"github.com/ayhanozemre/fs-shadow/watcher"
	log "github.com/sirupsen/logrus"
)

func main() {
	// not completed
	log.SetLevel(log.DebugLevel)
	tw, _, err := watcher.NewFSWatcher("/tmp/root")

	if err == nil {
		tw.PrintTree("INIT TREE")

		go func() {
			ch := tw.GetEvents()
			for {
				select {
				case p := <-ch:
					log.Debug(p.UUID)
				}
			}
		}()

		done := make(chan bool)
		<-done
	} else {
		log.Panic(err)
	}
	tw.Close()

}
