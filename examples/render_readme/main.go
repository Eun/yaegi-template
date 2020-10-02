// read README.md.tmpl and find all ```go ``` snippets and run them
package main

import (
	"os"

	yaegi_template "github.com/Eun/yaegi-template"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func main() {
	template := yaegi_template.MustNew(interp.Options{}, stdlib.Symbols)

	in, err := os.Open("README.md.tmpl")
	if err != nil {
		panic(err)
	}
	defer in.Close()

	out, err := os.Create("README.md")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	template.MustParse(in)
	template.MustExec(out, nil)
}
