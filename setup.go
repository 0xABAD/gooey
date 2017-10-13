// +build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	locals := []string {
		`"./filewatch"`,
		`"./websocket"`,
	}
	gopaths := []string {
		`"github.com/0xABAD/gooey/filewatch"`,
		`"github.com/0xABAD/gooey/websocket"`,
	}

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "favicon":
			if len(os.Args) < 4 {
				fmt.Println("USAGE: setup favicon FILE PACKAGE_NAME")
			} else {
				packageName = os.Args[3]
				favgen(os.Args[2], packageName)
			}

		case "local":
			reimport(gopaths, locals)

		case "gopath":
			reimport(locals, gopaths)

		default:
			fmt.Println("Unrecognized setup command:", os.Args[1])
		}
	} else {
		fmt.Println("No sub command given.")
	}
}

const FAVICON_GO = `package %s

// This file is generated.  Do not modify.

// Favicon with standard base64 encoding.
const FAVICON = "%s"
`

// Takes the web/favicon.ico file, encodes it to a base 64 string,
// and outputs to generated server/favicon.go file to be embedded
// in the binary.
func favgen(faviconPath, packageName string) {
	fav, err := ioutil.ReadFile(faviconPath)
	if err != nil {
		log.Fatalf("Failed to read %s file -- %s", faviconPath, err)
	}
	out, err := os.Create("favicon.go")
	if err != nil {
		log.Fatalln("Failed to create output favicon.go file --", err)
	}
	defer out.Close()

	str := base64.StdEncoding.EncodeToString(fav)
	fmt.Fprintf(out, FAVICON_GO, packageName, str)
}

// Changes the imports in file, gooey.go, to have the import names
// in the 'from' slice to become the corresponding name in the 'to'
// slice.  In other words, if an import in gooey.go is equal to
// 'from[i]' then it will be changed to the value in 'to[i]'.  After
// the imports are changed then changes are written back to gooey.go.
func reimport(from, to []string) {
	if len(from) != len(to) {
		log.Fatalln("from and to slices in reimport don't have the same length.")
	}

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, "gooey.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatalln("Failed to parse gooey.go.  Is setup being run from gooey's package directory? --", err)
	}

	for _, s := range parsed.Imports {
		for i, imp := range from {
			if s.Path.Value == imp {
				s.Path.Value = to[i]
			}
		}
	}

	f, err := os.Create("gooey.go")
	if err != nil {
		log.Fatalln("Could not modify gooey.go to localize imports --", err)
	}
	format.Node(f, fset, parsed)
}
