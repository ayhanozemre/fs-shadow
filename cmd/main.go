package main

import (
	"fmt"
	"github.com/rjeczalik/notify"
	"golang.org/x/sys/unix"
	"log"
)

func main() {
	//tw, err := watcher.NewFSPathWatcher("/home/wade/Desktop/TransferChain")
	//tw, err := watcher.NewVirtualPathWatcher("/home/wade/Desktop/TransferChain")

	c := make(chan notify.EventInfo, 1)
	go func() {
		moves := make(map[uint32]struct {
			From   string
			To     string
			Rename string
		})
		for {
			ei := <-c
			cookie := ei.Sys().(*unix.InotifyEvent).Cookie
			info := moves[cookie]
			switch ei.Event() {
			case notify.InMovedFrom:
				info.From = ei.Path()
			case notify.InMovedTo:
				info.To = ei.Path()
			case notify.Rename:
				info.Rename = ei.Path()
			}

			moves[cookie] = info

			if cookie != 0 && info.From != "" && info.To != "" {
				log.Println("File:", info.From, "was renamed to", info.To)
				delete(moves, cookie)
			} else if cookie != 0 && info.From != "" && info.Rename != "" {
				log.Println("File:", info.From, "moved to another destination")
			} else {
				fmt.Println("X", ei)
			}
		}
	}()

	err := notify.Watch("/tmp/fs-shadow-test/...",
		c,
		notify.All,
		notify.InMovedFrom,
		notify.InMovedTo)
	if err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	done := make(chan bool)
	<-done
	/*
		if err == nil {
			tw.PrintTree("INIT TREE")
			done := make(chan bool)
			<-done
		} else {
			logrus.Panic(err)
		}
		tw.Close()

	*/

}
