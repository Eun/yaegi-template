package yaegi_template

import (
	"io"
	"testing"

	"bytes"

	"bufio"

	"errors"

	"reflect"

	"fmt"
	"strconv"

	"os"
	"path/filepath"

	"io/ioutil"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func equalError(a, b error) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Error() == b.Error()
}

func TestExec(t *testing.T) {
	tests := []struct {
		Name         string
		Options      interp.Options
		Use          []interp.Exports
		Template     string
		ExpectBuffer string
		ExpectError  error
	}{
		{
			"Hello Yaegi",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<html><$fmt.Print("Hello Yaegi")$></html>`,
			`<html>Hello Yaegi</html>`,
			nil,
		},
		{
			"Func",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<html><$func Foo(text string) {
	fmt.Printf("Hello %s", text)
}$>
<p><$Foo("Yaegi")$></p>
</html>`,
			`<html>
<p>Hello Yaegi</p>
</html>`,
			nil,
		},

		{
			"Error",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$ Hello $>`,
			"",
			errors.New(`1:29: undefined: Hello`),
		},
		{
			"Import",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$import "net/url"$><$fmt.Print(url.PathEscape("Hello World"))$>`,
			"Hello%20World",
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			template := MustNew(test.Options, test.Use...).MustParseString(test.Template)
			var buf bytes.Buffer
			n, err := template.Exec(&buf, nil)
			if !equalError(test.ExpectError, err) {
				t.Fatalf("expected %#v, got %#v", test.ExpectError, err.Error())
			}
			if test.ExpectBuffer != buf.String() {
				t.Fatalf("expected %#v, got %#v", test.ExpectBuffer, buf.String())
			}
			if l := len(test.ExpectBuffer); l != n {
				t.Fatalf("expected %d, got %d", l, n)
			}
		})
	}
}

func TestExecWithContext(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	tests := []struct {
		Name                   string
		Options                interp.Options
		Use                    []interp.Exports
		Template               string
		ContextRun1            interface{}
		ExpectContextAfterRun1 interface{}
		ExpectBufferRun1       string
		ContextRun2            interface{}
		ExpectContextAfterRun2 interface{}
		ExpectBufferRun2       string
		ExpectError            error
	}{
		{
			"Struct",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`Hello <$fmt.Printf("%s (%d)", context.Name, context.ID)$>`,
			User{10, "Yaegi"},
			User{10, "Yaegi"},
			`Hello Yaegi (10)`,
			User{11, "Joe"},
			User{11, "Joe"},
			`Hello Joe (11)`,
			nil,
		},
		{
			"PtrStruct",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`Hello <$fmt.Printf("%s (%d)", context.Name, context.ID)$>`,
			&User{10, "Yaegi"},
			&User{10, "Yaegi"},
			`Hello Yaegi (10)`,
			&User{11, "Joe"},
			&User{11, "Joe"},
			`Hello Joe (11)`,
			nil,
		},
		{
			"Map",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$fmt.Printf("%d %s", context["Foo"], context["Bar"])$>`,
			map[string]interface{}{"Foo": 10, "Bar": "Yaegi"},
			map[string]interface{}{"Foo": 10, "Bar": "Yaegi"},
			`10 Yaegi`,
			map[string]interface{}{"Foo": 11, "Bar": "Joe"},
			map[string]interface{}{"Foo": 11, "Bar": "Joe"},
			`11 Joe`,
			nil,
		},
		{
			"Package",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$
		package main
		func main() {
		fmt.Printf("%d %s", context["Foo"], context["Bar"])
		}
		$>`,
			map[string]interface{}{"Foo": 10, "Bar": "Yaegi"},
			map[string]interface{}{"Foo": 10, "Bar": "Yaegi"},
			`10 Yaegi`,
			map[string]interface{}{"Foo": 11, "Bar": "Joe"},
			map[string]interface{}{"Foo": 11, "Bar": "Joe"},
			`11 Joe`,
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			template := MustNew(test.Options, test.Use...).MustParseString(test.Template)
			var buf bytes.Buffer
			n, err := template.Exec(&buf, test.ContextRun1)
			if !equalError(test.ExpectError, err) {
				t.Fatalf("expected %#v, got %#v", test.ExpectError, err)
			}
			if test.ExpectBufferRun1 != buf.String() {
				t.Fatalf("expected %#v, got %#v", test.ExpectBufferRun1, buf.String())
			}
			if !reflect.DeepEqual(test.ExpectContextAfterRun1, test.ContextRun1) {
				t.Fatalf("expected %#v, got %#v", test.ExpectContextAfterRun1, test.ContextRun1)
			}

			if l := len(test.ExpectBufferRun1); l != n {
				t.Fatalf("expected %d, got %d", l, n)
			}

			// run again with the second context
			buf.Reset()
			n, err = template.Exec(&buf, test.ContextRun2)
			if !equalError(test.ExpectError, err) {
				t.Fatalf("expected %#v, got %#v", test.ExpectError, err)
			}
			if test.ExpectBufferRun2 != buf.String() {
				t.Fatalf("expected %#v, got %#v", test.ExpectBufferRun2, buf.String())
			}
			if !reflect.DeepEqual(test.ExpectContextAfterRun2, test.ContextRun2) {
				t.Fatalf("expected %#v, got %#v", test.ExpectContextAfterRun2, test.ContextRun2)
			}

			if l := len(test.ExpectBufferRun2); l != n {
				t.Fatalf("expected %d, got %d", l, n)
			}
		})
	}
}

func TestSkipIdent(t *testing.T) {
	tests := []struct {
		Name             string
		Token            string
		Input            []byte
		ExpectBuffer     []byte
		ExpectReadError  error
		ExpectWriteError error
	}{
		{"Single",
			"{",
			[]byte("Hello{World"),
			[]byte("Hello"),
			nil,
			nil,
		},
		{"Double",
			"{%",
			[]byte("Hello{%World"),
			[]byte("Hello"),
			nil,
			nil,
		},
		{"Double (same rune)",
			"{{",
			[]byte("Hello{{{World"),
			[]byte("Hello"),
			nil,
			nil,
		},
		{"Only find the first",
			"{%",
			[]byte("Hello{%World{%Bye"),
			[]byte("Hello"),
			nil,
			nil,
		},
		{"50% invalid",
			"{%",
			[]byte("Foo{{Bar{%Baz"),
			[]byte("Foo{{Bar"),
			nil,
			nil,
		},
		{"On start",
			"{%",
			[]byte("{%Baz"),
			[]byte(""),
			nil,
			nil,
		},
		{"On end",
			"{%",
			[]byte("Baz{%"),
			[]byte("Baz"),
			nil,
			nil,
		},
		{"Nothing at all",
			"{%",
			[]byte("Bar"),
			[]byte("Bar"),
			io.EOF,
			nil,
		},
		{
			"",
			"}%",
			[]byte("Hello}}%"),
			[]byte("Hello}"),
			nil,
			nil,
		},
	}
	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			var buf bytes.Buffer
			n, rerr, werr := skipIdent([]rune(test.Token), bufio.NewReader(bytes.NewReader(test.Input)), &buf)
			if test.ExpectReadError != rerr {
				t.Fatalf("expected %#v, got %#v", test.ExpectReadError, rerr)
			}
			if test.ExpectWriteError != werr {
				t.Fatalf("expected %#v, got %#v", test.ExpectWriteError, werr)
			}
			if !bytes.Equal(test.ExpectBuffer, buf.Bytes()) {
				t.Fatalf("expected %#v (%#v), got %#v (%#v)", test.ExpectBuffer, string(test.ExpectBuffer), buf.Bytes(), buf.String())
			}
			if len(test.ExpectBuffer) != n {
				t.Fatalf("expected %#v, got %#v", len(test.ExpectBuffer), n)
			}
		})
	}
}

func TestTemplate_MustParse(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols).
		MustParse(bytes.NewBufferString(`Hello <$ fmt.Print("Yaegi") $>`))

	var buf bytes.Buffer
	if _, err := template.Exec(&buf, nil); err != nil {
		t.Fatal("exec failed")
	}
	if buf.String() != "Hello Yaegi" {
		t.Fatalf(`expected "Hello Yaegi", got %#v`, buf.String())
	}
}

func TestMultiExec(t *testing.T) {
	type MessageContext struct {
		Message string
	}

	var templates []*Template

	for i := 0; i < 3; i++ {
		t := MustNew(interp.Options{}, stdlib.Symbols)
		t.MustParseString(`<$fmt.Printf("Hello ` + strconv.Itoa(i) + ` %s", context.Message)$>`)
		templates = append(templates, t)
	}

	run := func(i int) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			templateIndex := i % len(templates)
			msg := MessageContext{Message: strconv.Itoa(i)}
			expect := fmt.Sprintf("Hello %d %d", templateIndex, i)
			var buf bytes.Buffer
			templates[templateIndex].MustExec(&buf, msg)
			if expect != buf.String() {
				t.Fatalf(`expected %s, got %#v`, expect, buf.String())
			}
		}
	}

	for i := 0; i < 1000; i++ {
		t.Run(strconv.Itoa(i), run(i))
	}
}

func TestFmtSprintf(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$fmt.Printf(fmt.Sprintf("Hello %s", "World"))$>`)
	var buf1 bytes.Buffer
	template.MustExec(&buf1, nil)
	if buf1.String() != "Hello World" {
		t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
	}
}

func TestPanic(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$panic("Oh no")$>`)
	var buf1 bytes.Buffer
	_, err := template.Exec(&buf1, nil)
	if !equalError(err, errors.New("Oh no")) {
		t.Fatalf("expected Oh no, got %v", err)
	}
}

func TestNoStartOrEnd(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.StartTokens = []rune{}
	template.EndTokens = []rune{}
	template.MustParseString(`fmt.Printf(fmt.Sprintf("Hello %s", "World"))`)
	var buf1 bytes.Buffer
	template.MustExec(&buf1, nil)
	if buf1.String() != "Hello World" {
		t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
	}
}

func TestImport(t *testing.T) {
	// create sample package
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unable to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmp)

	srcPath := filepath.Join(tmp, "src", "world")
	if err = os.MkdirAll(srcPath, 0777); err != nil {
		t.Fatalf("unable to create temp dir /src/world: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(srcPath, "world.go"), []byte(`
package world
func World() string {
    return "World"
}`), 0777)
	if err != nil {
		t.Fatalf("unable to create world.go: %v", err)
	}

	t.Run("simple", func(t *testing.T) {
		template := MustNew(interp.Options{
			GoPath: tmp,
		}, stdlib.Symbols)
		template.StartTokens = []rune{}
		template.EndTokens = []rune{}
		template.MustParseString(`
import "world"
fmt.Printf(fmt.Sprintf("Hello %s", world.World()))`)

		var buf1 bytes.Buffer
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}

		buf1.Reset()
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}
	})

	t.Run("multi import", func(t *testing.T) {
		template := MustNew(interp.Options{
			GoPath: tmp,
		}, stdlib.Symbols)
		template.StartTokens = []rune{}
		template.EndTokens = []rune{}
		template.MustParseString(`
import (
"fmt"
"world"
)
fmt.Printf(fmt.Sprintf("Hello %s", world.World()))`)

		var buf1 bytes.Buffer
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}

		buf1.Reset()
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}
	})

	t.Run("multi import", func(t *testing.T) {
		template := MustNew(interp.Options{
			GoPath: tmp,
		}, stdlib.Symbols)
		template.StartTokens = []rune{}
		template.EndTokens = []rune{}
		template.MustParseString(`
import (
"fmt"
)

import (
"world"
)
fmt.Printf(fmt.Sprintf("Hello %s", world.World()))`)

		var buf1 bytes.Buffer
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}

		buf1.Reset()
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}
	})

	t.Run("alias import", func(t *testing.T) {
		template := MustNew(interp.Options{
			GoPath: tmp,
		}, stdlib.Symbols)
		template.StartTokens = []rune{}
		template.EndTokens = []rune{}
		template.MustParseString(`
import (
"fmt"
w "world"
)
fmt.Printf(fmt.Sprintf("Hello %s", w.World()))`)

		var buf1 bytes.Buffer
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}

		buf1.Reset()
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}
	})

	t.Run("alias dot import", func(t *testing.T) {
		template := MustNew(interp.Options{
			GoPath: tmp,
		}, stdlib.Symbols)
		template.StartTokens = []rune{}
		template.EndTokens = []rune{}
		template.MustParseString(`
import (
"fmt"
. "world"
)
fmt.Printf(fmt.Sprintf("Hello %s", World()))`)

		var buf1 bytes.Buffer
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}

		buf1.Reset()
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}
	})

	t.Run("package", func(t *testing.T) {
		template := MustNew(interp.Options{
			GoPath: tmp,
		}, stdlib.Symbols)
		template.StartTokens = []rune{}
		template.EndTokens = []rune{}
		template.MustParseString(`
package main
import "world"
func main() {
fmt.Printf(fmt.Sprintf("Hello %s", world.World()))
}`)

		var buf1 bytes.Buffer
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}

		buf1.Reset()
		template.MustExec(&buf1, nil)
		if buf1.String() != "Hello World" {
			t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
		}
	})
}

func TestImplicitReturn(t *testing.T) {
	ctx := map[string]interface{}{
		"Bool":   false,
		"Int":    1,
		"Uint":   uint(1),
		"Float":  1.2,
		"String": "Foo",
		"Func":   func() {},
	}
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$context["Bool"]$> <$context["Int"]$> <$context["Uint"]$> <$context["Float"]$> <$context["String"]$> <$context["Func"]$>`)
	var buf bytes.Buffer
	template.MustExec(&buf, ctx)

	if buf.String() != `false 1 1 1.2 Foo ` {
		t.Fatalf(`expected "false 1 1 1.2 Foo ", got %#v`, buf.String())
	}
}
