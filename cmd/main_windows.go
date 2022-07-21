package main

import (
	"fmt"
	"github.com/ayhanozemre/fs-shadow/watcher"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	log.SetLevel(log.DebugLevel)
	//tw, err := watcher.NewVirtualPathWatcher("/home/wade/Desktop/TransferChain")
	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			fmt.Fprintf(os.Stderr, "Exception: %v\n", err)
			time.Sleep(20 * time.Second)
		}
	}()
	tw, err := watcher.NewFSWatcher("TransferChain")

	if err == nil {
		tw.PrintTree("INIT TREE")
		done := make(chan bool)
		<-done
	} else {
		log.Debug(err)
		time.Sleep(3 * time.Second)
	}
	tw.Close()
}
