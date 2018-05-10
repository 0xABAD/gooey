package gooey

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/0xABAD/gooey/filewatch"
	"github.com/0xABAD/gooey/websocket"
)

// App provides a means sending and receiving websocket messages to connected clients.
type App interface {
	// Start is called by a gooey Server whenever a client connects to the server.  It
	// will be called once for each connecting client so any state to the connection
	// should be local to the function and any global state shared between all connections
	// should within App (or global).
	//
	// All messages to App are passed into the incoming channel and any messages to the
	// client should be passed on the outgoing channel.  The contents of byte slice passed
	// into the incoming channel are entirely up what's passed from the client as the
	// gooey Server passes the msg as is received.  Note that the gooey Server processes
	// all websocket messages as text messages.  The content passed to the outgoing
	// channel will be encoded as JSON before being sent on the websocket.  The message
	// on the client can be received by overriding the gooey.OnMessage function.  See
	// the documentation in gooey.js for more information on processing client side code.
	//
	// The closed channel will be closed when the connection with client has been closed
	// (i.e. the user closes the browser tab).  This allows to perform any clean up
	// that the App needs to perform.
	Start(closed <-chan struct{}, incoming <-chan []byte, outgoing chan<- interface{})
}

// Server represents an active server connection that can listen to incoming connecting
// clients.
type Server struct {
	// The IP address that the server should listen on.  It should be in the form of
	// "127.0.0.1:8080" such as accepted by the net and net/http packages.  If Addr
	// is the empty string then Server will listen on "127.0.0.1:" which will allow
	// the OS to select a random port for the connection.
	Addr string

	// The directory of where Server should serve web files from.  If WebServeDir is
	// the empty string then Server will serve files from a created temporary directory.
	// Server will generate index.html and favicon.ico files, dependent on the values
	// set for IndexHtml and FavIcon fields, and place them inside this temp directory
	// to be served.
	WebServeDir string

	// The index.html file that will be served to incoming client connections.  Note that
	// this is the contents of the index file, not the path to some index.html file.  It
	// it expected that these contents are embedded within the executable or loaded by
	// other means.  The contents of the string will be written to a temporary index.html
	// file and served to clients unless WebServeDir is not the empty string.  If this
	// field is the empty string then Server will use a default index.html file.
	IndexHtml string

	// The contents of the favicon.ico that is encoded in base64.  Like IndexHtml, these
	// are expected to be the actual favicon contents that will be written to temp file
	// to be served unless WebServeDir is set.  If this field is the empty string then
	// a default gooey favicon will be used.
	FavIcon string

	// Specifies a directory whose contents will be watched (recursively) for changes and
	// when a change is detected then a special message will be sent to the client to
	// reload the page contents.  If this field is the empty string then no hot reloading
	// will occur.
	//
	// There are special rules for which files be sent for the hot reload.  First, if
	// one of the files that has changed is named body.html then the contents of that
	// file will replace the body of the current document loaded in the client.  If body.html
	// is removed then the body of the document will be replaced with an empty <div> tag.
	// If a CSS or a Javascript file is added or changed then all known CSS or Javascript
	// content will be concatenated together and replace a gooey specially constructed
	// <style> or <script> tag that is appended at the end of the document <head> element,
	// if it exist.  If a CSS or a Javascript file is removed then the contents of the
	// file will be omitted from the concatenated content.
	//
	// Javascript files have an additional rule that allows enforcement of ordering of
	// Javascript content that will be embedded in the document.  If a Javascript file has a
	// number after a dot in the file name but before the .js extenstion then that number
	// declares the ordering amongst other ordered Javascript files which will also be inserted
	// before other non ordered watched javascript files.  For example, suppose we have three
	// Javascript files: baz.js, bar.3.js, and foo.0.js, and suppose the current <head>
	// tag is as follows:
	//
	// <head>
	//   <script src="/somescript.js"></script>
	//   <link rel="stylesheet" href="/somestyle.css">
	// </head>
	//
	// Then the hot reload output will be:
	//
	// <head>
	//   <script src="/somescript.js"></script>
	//   <link rel="stylesheet" href="/somestyle.css">
	//   <script id="gooey-reload-js-content">
	//   /* concatenated content of foo.0.js, bar.3.js, and then baz.js */
	//   </script>
	// </head>
	ReloadWatchDir string

	// If ReloadWatchDir is not the empty then  gooey will ignore all files
	// that match any of these patterns.  That match algorithm used is the
	// same as specified in path.Match.
	ReloadIgnorePatterns []string

	// If this field is set to true then the server will not automatically shutdown after
	// the last client connection to the server is closed.
	NoAutoShutdown bool

	// If this field is set to true then the server will not open a browser tab in the
	// users default browser on server start.  It should be noted that if this field
	// is set to true then it might be wise to assign a custom address to the Addr
	// field so one knows how to connect to the server.
	NoAutoOpen bool

	// A logger for the server to post informational to, essentially enabling a verbose
	// mode.  If set to nil then no info messages will be posted.
	InfoLog *log.Logger

	// A logger for the server to post runtime error messages to.  If set to nil then
	// no error messages will be posted.  Note that the errors posted to this logger
	// are not the same that returned from the Start method, as those are initialization
	// errors and the server can not start.  These errors are runtime errors that are
	// not necessarily fatal to the server (e.g. a websocket WritMessage error).
	ErrorLog *log.Logger

	// A channel, if non nil, will be passed any runtime errors that are encountered
	// during server operation.  These errors are the exact same that are passed to
	// ErrorLog.  Note that these two fields are not mutually exclusive as any error
	// encountered will be written to the error log and this channel.
	ErrorC chan<- error
}

// Start the server and allow incoming client connections. If an intialization error
// occurs then the start fails and that error is returned, otherwise, the server is
// started and this call blocks until there are no more clients connected (if
// NoAutoShutdown is false) or the done channel is closed.
func (server *Server) Start(done <-chan struct{}, app App) error {
	addr := "127.0.0.1:"
	if server.Addr != "" {
		addr = server.Addr
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("Failed to create net.listener -- %s", err)
	}
	defer listener.Close()

	dir, err := ioutil.TempDir("", "gooey_server")
	if err != nil {
		return fmt.Errorf("Failed to create a temporary gooey_server directory -- %s\n", err)
	}
	defer os.RemoveAll(dir)

	redirect, err := buildTempFile(dir, "redirect.html", REDIRECT, listener)
	if err != nil {
		return err
	}
	defer redirect.close()

	index, err := createTempFile(dir, "index.html")
	if err != nil {
		return err
	} else {
		page := INDEX
		if server.IndexHtml != "" {
			page = server.IndexHtml
		}
		_, err := index.WriteString(page)
		if err != nil {
			index.close()
			return fmt.Errorf("Failed to write temp index.html file -- %s\n", err)
		}
	}
	defer index.close()

	favstr := FAVICON
	if server.FavIcon != "" {
		favstr = server.FavIcon
	}
	favicon, err := base64.StdEncoding.DecodeString(favstr)
	if err != nil {
		server.errorln("Failed to decode favicon --", err)
	} else {
		// If this fails then we move on as it's not the end of the world
		// if we are missing the favicon.
		favpath := filepath.Join(dir, "favicon.ico")
		if err := ioutil.WriteFile(favpath, favicon, os.FileMode(0644)); err != nil {
			server.errorln("Failed to create favicon in temp directory --", err)
		}
	}

	http.HandleFunc("/gooeynewtab", func(w http.ResponseWriter, r *http.Request) {
		exec.Command(BROWSE, redirect.name()).Start()
	})

	if server.WebServeDir != "" {
		http.Handle("/", http.FileServer(http.Dir(server.WebServeDir)))
	} else {
		http.Handle("/", http.FileServer(http.Dir(dir)))
	}

	var (
		onOpen   = make(chan *websocket.Conn)
		shutdown = make(chan struct{})
	)

	go server.monitorClients(done, onOpen, shutdown, app)
	http.HandleFunc("/gooeywebsocket", server.handleWebsocket(onOpen))
	if !server.NoAutoOpen {
		exec.Command(BROWSE, redirect.name()).Start()
	}
	go http.Serve(listener, nil)

	<-shutdown

	return nil
}

func (s *Server) handleWebsocket(onOpen chan<- *websocket.Conn) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ws := websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     func(r *http.Request) bool { return true },
		}
		c, err := ws.Upgrade(w, r, nil)
		if err != nil {
			s.errorln("Failed to upgrade websocket connection -- ", err)
		} else {
			onOpen <- c
		}
	}
}

func (server *Server) monitorClients(done <-chan struct{}, onOpen <-chan *websocket.Conn, shutdown chan<- struct{}, app App) {
	var (
		connections = 0
		onClose     = make(chan struct{})
		noMoreConns = make(chan struct{})
		noMoreTimer *time.Timer
	)

	if server.NoAutoShutdown {
		for {
			select {
			case <-done:
				close(shutdown)
				return
			case conn := <-onOpen:
				go server.connect(conn, done, onClose, app)
			}
		}
	}

	for {
		select {
		case conn := <-onOpen:
			open := func() {
				connections++
				server.infoln("Connection opened -- count", connections)
				go server.connect(conn, done, onClose, app)
			}
			// In case the user is spamming the refresh button on the browser we want
			// stop the timer as the connection is being reopened.  This stops a case
			// where a timer from an earlier refresh is fired at the right time during
			// another refresh and connections == 0 which will then inadvertently
			// shut down the server.
			if noMoreTimer == nil {
				open()
			} else {
				if noMoreTimer.Stop() {
					noMoreTimer = nil
					open()
				}
			}

		case <-onClose:
			connections--
			server.infoln("Connection closed -- count", connections)
			// We want to give some time before shutting down altogether to
			// handle the case where there is only one connection and the
			// page is being refreshed, which causes the connection to close
			// and reopen immediately.
			if noMoreTimer == nil {
				noMoreTimer = time.AfterFunc(500*time.Millisecond, func() { noMoreConns <- struct{}{} })
			}

		case <-noMoreConns:
			if connections < 0 {
				// If we get here then the synchronization is broken.
				panic("Number of connections dropped below zero")
			} else if connections == 0 {
				server.infoln("Shutting down gooey web server")
				close(shutdown)
				return
			} else {
				// Suppose the user has two tabs open, closes one, and eventually the
				// noMoreTimer fires.  In this case, we reach this point here and must
				// set the timer to nil in case the user opens another tab.  If not, then
				// noMoreTimer.Stop() returns false and the connection doesn't open.
				noMoreTimer = nil
			}
		}
	}
}

func (server *Server) connect(conn *websocket.Conn, done <-chan struct{}, onClose chan<- struct{}, app App) {
	var (
		stop     = make(chan struct{})
		reload   = make(chan interface{})
		incoming = make(chan []byte)
		outgoing = make(chan interface{})
	)

	go app.Start(stop, incoming, outgoing)

	// This is pretty inefficient to watch the reload directory for each websocket
	// connection but we'll allow it as it is intended for development where it is
	// expected that only one or two tabs may be open at a time.
	if server.ReloadWatchDir != "" {
		interval := 1 * time.Second
		updates, err := filewatch.Watch(done, server.ReloadWatchDir, true, &interval)

		if err != nil {
			server.errorln("Could not watch web files --", err)
		} else {
			go func() {
				var (
					buf bytes.Buffer
					js  = make(map[string]string)
					css = make(map[string]string)
				)
				for {
					select {
					case <-done:
						return
					case us := <-updates:
						reload <- server.reloadWebContent(us, css, js, buf)
					}
				}
			}()
		}
	}

	go (func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					server.infoln("Client closing connection")
					close(stop)
					return
				} else {
					ok := true
					select {
					case _, ok = <-stop:
					case _, ok = <-done:
					default:
						server.errorln("ReadMessage error --", err)
					}
					if !ok {
						return
					}
				}
			} else {
				incoming <- msg
			}
		}
	})()

	send := func(out interface{}) {
		if text, err := json.Marshal(out); err != nil {
			server.errorln("Failed to marshal JSON message --", err)
		} else if err := conn.WriteMessage(websocket.TextMessage, text); err != nil {
			server.errorln("WriteMessage failed to send message --", err)
		}
	}

	for {
		select {
		case <-stop:
			server.infoln("Shutting down websocket connection")
			conn.Close()
			onClose <- struct{}{}
			return

		case <-done:
			msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
			if err := conn.WriteMessage(websocket.CloseMessage, msg); err != nil {
				server.errorln("WriteMessage error --", err)
			} else {
				server.infoln("Writing websocket close message")
			}
			conn.Close()
			onClose <- struct{}{}
			return

		case content := <-outgoing:
			send(content)

		case content := <-reload:
			server.infoln("Reloading web content")
			send(struct {
				GooeyMessage string
				GooeyContent interface{}
			}{
				GooeyMessage: "gooey-server-reload-content",
				GooeyContent: content,
			})
		}
	}
}

type contentUpdate struct {
	Body, Javascript, CSS string
}

func (s *Server) reloadWebContent(updates []filewatch.Update, css, js map[string]string, buf bytes.Buffer) contentUpdate {
	var update contentUpdate

	check := func(u filewatch.Update, m map[string]string) (hasUpdate bool) {
		if u.WasRemoved {
			hasUpdate = true
			delete(m, u.AbsPath)
		} else {
			if content, err := ioutil.ReadFile(u.AbsPath); err != nil {
				// This can occur for files that are created by other programs,
				// such as a text editor, which may create backup files and
				// delete them before the next filewatch update is received.
				if os.IsNotExist(err) {
					hasUpdate = true
					delete(m, u.AbsPath)
				} else {
					s.errorln("Failed to read", u.AbsPath, "--", err)
				}
			} else {
				hasUpdate = true
				m[u.AbsPath] = string(content)
			}
		}
		return
	}

	var jsUpdate, cssUpdate bool

	for _, u := range updates {
		ignore := false
		for _, pattern := range s.ReloadIgnorePatterns {
			match, err := path.Match(pattern, filepath.Base(u.AbsPath))
			if err != nil {
				s.errorln("Failed to match pattern", pattern, "for path", u.AbsPath, "--", err)
			} else if match {
				ignore = true
				break
			}
		}
		if ignore {
			continue
		}

		if strings.HasSuffix(u.AbsPath, "body.html") {
			if u.WasRemoved {
				update.Body = "<div></div>"
			} else {
				if body, err := ioutil.ReadFile(u.AbsPath); err != nil {
					s.errorln("Failed to read", u.AbsPath, "--", err)
				} else {
					update.Body = string(body)
				}
			}
		}
		if strings.HasSuffix(u.AbsPath, ".js") {
			jsUpdate = check(u, js)
		}
		if strings.HasSuffix(u.AbsPath, ".css") {
			cssUpdate = check(u, css)
		}
	}

	if cssUpdate {
		buf.Reset()
		for _, c := range css {
			buf.WriteString(c)
		}
		update.CSS = buf.String()
	}

	// combine and sort
	if jsUpdate {
		buf.Reset()

		numbered := make(map[int]string, len(js))
		nonsorted := make([]string, 0, len(js))

		for path, _ := range js {
			sp := strings.Split(filepath.Base(path), ".")
			ln := len(sp)
			if ln >= 3 {
				n := sp[ln-2]
				if len(n) > 1 {
					n = strings.TrimLeft(n, "0")
				}
				if i, err := strconv.ParseInt(n, 10, 32); err == nil {
					numbered[int(i)] = path
				} else {
					nonsorted = append(nonsorted, path)
				}
			} else {
				nonsorted = append(nonsorted, path)
			}
		}

		idx, sorted := 0, make([]int, len(numbered))
		for i, _ := range numbered {
			sorted[idx] = i
			idx++
		}
		sort.Ints(sorted)

		for _, n := range sorted {
			path := numbered[n]
			buf.WriteString(js[path])
		}
		for _, path := range nonsorted {
			buf.WriteString(js[path])
		}
		update.Javascript = buf.String()
	}

	return update
}

func (s *Server) infoln(args ...interface{}) {
	if s.InfoLog != nil {
		s.InfoLog.Output(2, fmt.Sprintln(args...))
	}
}

func (s *Server) errorln(args ...interface{}) {
	if s.ErrorLog != nil {
		s.ErrorLog.Output(2, fmt.Sprintln(args...))
	}
	if s.ErrorC != nil {
		go func() {
			s.ErrorC <- fmt.Errorf(fmt.Sprintln(args...))
		}()
	}
}

func createTempFile(dir, name string) (*tempFile, error) {
	file, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return nil, fmt.Errorf("Failed to create a temporary file, %s -- %s\n", name, err)
	}
	return (*tempFile)(file), nil
}

func buildTempFile(dir, name, tpl string, data interface{}) (*tempFile, error) {
	file, err := createTempFile(dir, name)
	if err != nil {
		return nil, err
	}
	return file, file.buildTemplate(tpl, data)
}

type tempFile os.File

func (f *tempFile) close() {
	file := (*os.File)(f)
	if file != nil {
		file.Close()
	}
}

func (f *tempFile) name() string {
	file := (*os.File)(f)
	return file.Name()
}

func (f *tempFile) WriteString(s string) (n int, err error) {
	file := (*os.File)(f)
	return file.WriteString(s)
}

func (f *tempFile) Write(bs []byte) (n int, err error) {
	file := (*os.File)(f)
	return file.Write(bs)
}

func (file *tempFile) buildTemplate(tpl string, data interface{}) error {
	name := file.name()
	temp, err := template.New(name).Parse(tpl)

	if err != nil {
		return fmt.Errorf("Failed to parse template, %s -- %s\n", name, err)
	}
	if err := temp.Execute(file, data); err != nil {
		return fmt.Errorf("Failed to execute template, %s -- %s\n", name, err)
	}
	file.close()
	return nil
}
