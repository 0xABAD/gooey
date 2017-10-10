Gooey
=====

A framework to create single executable web apps.

Overview
--------

Gooey provides a base to create a single executable that displays a GUI through
the user's default browser.  It is designed to automatically open a browser tab
when double clicked, communicate through websockets, and automatically shutdown
when the user closes the last tab connected to the server.  Some use cases for
Gooey are simple tools that could benefit from a GUI or a debug dashboard for
a service, all without the heavy weight of something like Electron or having try
to use GUI bindings that are tied to a particular platform.

Quickstart
----------

First, if you haven't already, ```go get github.com/0xABAD/gooey```.  Then
create a new directory for a Go package and copy this code:

```
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
		app	   testApp
		server gooey.Server
		notify = make(chan os.Signal)
		done   = make(chan struct{})
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
```

Now run `go run test.go`.  You should see something like this:

![Gooey Demo](demo.gif)

Some things you may have noticed:

* A browser window was automatically opened
* The address shows a random port number assigned
* We see the message being pushed from the server once a second
* The page loaded was embedded within the executable
* When closing the tab the server was automatically shutdown

Of course, all of this functionality may be configured through the
`gooey.Server struct`.  See the 
[documentation](https://godoc.org/github.com/0xABAD/gooey) for more.

LICENSE
-------

Gooey is licensed under Zlib license but it does use
[gorilla/websocket](https://github.com/gorilla/websocket), while having a
permissive license, requires you redistibute its license whether redistibuting
in source or binary form.
