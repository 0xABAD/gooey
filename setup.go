// +build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "favicon":
			if len(os.Args) < 3 {
				fmt.Println("USAGE: setup favicon file")
				fmt.Println(`    Where "file" is the path of favicon.ico file to generate.`)
			} else {
				favgen(os.Args[2])
			}
		default:
			fmt.Println("Unrecognized setup command:", os.Args[1])
		}
	} 
}

const FAVICON_GO = `
package gooey

// This file is generated.  Do not modify.

// Favicon with standard base64 encoding.
const FAVICON = "%s"
`

// Takes the web/favicon.ico file, encodes it to a base 64 string,
// and outputs to generated server/favicon.go file to be embedded
// in the binary.
func favgen(faviconPath string) {
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
	fmt.Fprintf(out, FAVICON_GO, str)
}
