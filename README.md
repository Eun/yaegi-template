## yaegi-template
Use [yaegi](https://github.com/containous/yaegi) as a template engine.

Proof of concept only!
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
	func GreetUser(name string) {
		fmt.Fprintf(out, "Hello %s", name)
	}
$>

<p>
<$
	if context["LoggedIn"].(bool) {
		GreetUser(context["UserName"].(string))
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
		UserName: "Joe",
	})
}
```