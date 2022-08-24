package main

import (
	"fmt"
	"github.com/ayhanozemre/fs-shadow/watcher"
	log "github.com/sirupsen/logrus"
)

func main() {
	// not completed
	log.SetLevel(log.DebugLevel)

	tw, _, err := watcher.NewFSWatcher("C:\\TransferChain")

	if err == nil {
		tw.PrintTree("init")

		go func() {
			ch := tw.GetEvents()
			err := tw.GetErrors()
			for {
				select {
				case p := <-ch:
					log.Debug(fmt.Sprintf("Event-> Name:%s type:%s", p.Name, p.Type.String()))
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
