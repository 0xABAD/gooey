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
			if len(os.Args) < 4 {
				fmt.Println("USAGE: setup favicon FILE PACKAGE_NAME")
			} else {
				packageName := os.Args[3]
				favgen(os.Args[2], packageName)
			}

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
