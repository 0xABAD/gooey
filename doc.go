/*
Package gooey provides a web based GUI framework to create single executable web apps.

This package setups a server structure to produce an executable that launches a
GUI app through the user's default web browser where the server and client
communicate through a websocket.  The intent is to embed the web content within
the final executable in order to distribute a single executable to users that
can double-click on the executable and start interacting with the program.  The
use case is for simple tools that can use a simple GUI or debug dashboards for
server programs.  It allows for simple GUI based tools written in Go without the
heavy weight of an Electron app.

The following is an example of a minimalist gooey program:

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

After running this program notice that the user doesn't have to navigate to the
server's address in the web browser and that the tab opened automatically for
them.  Second, notice the port number is randomly assigned by the OS when gooey
setups the server.  Finally, see how the page updates when the messages from the
server are sent to the client.  This portrays gooey's client to server code that
is embedded within the embedded index.html page.

One other point is the use of waiting for the os.Kill and os.Interrupt signals
to cleanly shutdown the server.  While this isn't necessary it ensures a proper
clean up done by gooey to close current websocket connections and remove the
temporary directory that it creates on server start.  Also, note that by default
the user doesn't need to stop the server by signaling an interrupt; instead, the
user can simply close all open browser tabs connected to the server and the
server will shut itself down.

Note that all what you seen here can be configured through the Server struct.
Check the project's readme in its repository for a more in depth example that
uses the various configuration options.
*/
package gooey
