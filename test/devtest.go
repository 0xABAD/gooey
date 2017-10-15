// +build ignore

package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/0xABAD/gooey"
)

func main() {
	var (
		app    testApp
		notify = make(chan os.Signal)
		done   = make(chan struct{})
		server = gooey.Server{
			ReloadWatchDir:       "./web/",
			ReloadIgnorePatterns: []string{".#*"},
		}
	)
	signal.Notify(notify, os.Kill, os.Interrupt)
	go func() {
		<-notify
		close(done)
	}()
	server.Start(done, &app)
}

type testApp struct{}

func (a *testApp) Start(closed <-chan struct{}, incoming <-chan []byte, outgoing chan<- interface{}) {
	count := 0
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-closed:
			return
		case <-ticker.C:
			outgoing <- fmt.Sprintf("Message from server.  Count %d", count)
			count++
		}
	}
}
