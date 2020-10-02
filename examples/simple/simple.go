package main

import (
	"os"

	yaegi_template "github.com/Eun/yaegi-template"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
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
