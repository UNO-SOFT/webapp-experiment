// +build never

package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/securecookie"
)

func main() {
	fh, err := os.Create("keys.go")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(fh, `package main
import "encoding/base64"
var (
	cookieStoreKey, _ = base64.StdEncoding.DecodeString(`+"`%s`"+`)
	sessionStoreKey, _ = base64.StdEncoding.DecodeString(`+"`%s`"+`)
)`,
		base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)),
		base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(64)),
	)
	if err := fh.Close(); err != nil {
		log.Fatal(err)
	}
}
