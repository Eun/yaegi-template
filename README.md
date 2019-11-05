# yaegi-template [![Travis](https://img.shields.io/travis/Eun/yaegi-template.svg)](https://travis-ci.org/Eun/yaegi-template) [![Codecov](https://img.shields.io/codecov/c/github/Eun/yaegi-template.svg)](https://codecov.io/gh/Eun/yaegi-template) [![GoDoc](https://godoc.org/github.com/Eun/yaegi-template?status.svg)](https://godoc.org/github.com/Eun/yaegi-template) [![go-report](https://goreportcard.com/badge/github.com/Eun/yaegi-template)](https://goreportcard.com/report/github.com/Eun/yaegi-template)
Use [yaegi](https://github.com/containous/yaegi) as a template engine.

```go
package main

import (
	"os"

	"github.com/Eun/yaegi-template"
	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
)

func main() {
	template := yaegi_template.MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`
<html>
<$
	import "time"
	func GreetUser(name string) {
		fmt.Printf("Hello %s, it is %s", name, time.Now().Format(time.Kitchen))
	}
$>

<p>
<$
	if context.LoggedIn {
		GreetUser(context.UserName)
	}
$>
</p>
</html>
`)

	type Context struct {
		LoggedIn bool
		UserName string
	}

	template.MustExec(os.Stdout, &Context{
		LoggedIn: true,
		UserName: "Joe Doe",
	})
}
```