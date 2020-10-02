// read README.md and find all ```go ``` snippets and run them
package main

import (
	"io/ioutil"

	"os"

	yaegi_template "github.com/Eun/yaegi-template"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func main() {
	template := yaegi_template.MustNew(interp.Options{}, stdlib.Symbols)
	template.StartTokens = []rune("```go")
	template.EndTokens = []rune("```")

	f, err := os.Open("README.md")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	template.MustParse(f)
	// since we only care for errors throw throw the output away
	template.MustExec(ioutil.Discard, nil)
}
