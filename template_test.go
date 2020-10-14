package yaegi_template

import (
	"testing"

	"bytes"

	"fmt"
	"strconv"

	"os"
	"path/filepath"

	"io/ioutil"

	"github.com/stretchr/testify/require"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func TestExec(t *testing.T) {
	tests := []struct {
		Name         string
		Options      interp.Options
		Use          []interp.Exports
		Template     string
		ExpectOutput string
		ExpectError  string
	}{
		{
			"Hello Yaegi",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<html><$fmt.Print("Hello Yaegi")$></html>`,
			`<html>Hello Yaegi</html>`,
			"",
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
			"",
		},

		{
			"Error",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$ Hello $>`,
			"",
			`1:29: undefined: Hello`,
		},
		{
			"Import",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$import "net/url"$><$fmt.Print(url.PathEscape("Hello World"))$>`,
			"Hello%20World",
			"",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			template := MustNew(test.Options, test.Use...).MustParseString(test.Template)
			var buf bytes.Buffer
			n, err := template.Exec(&buf, nil)
			if test.ExpectError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.ExpectError)
			}
			require.Equal(t, test.ExpectOutput, buf.String())
			require.Len(t, test.ExpectOutput, n)
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
		ExpectOutputRun1       string
		ContextRun2            interface{}
		ExpectContextAfterRun2 interface{}
		ExpectOutputRun2       string
		ExpectError            string
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
			"",
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
			"",
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
			"",
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
			"",
		},
		{
			"Package (Indented)",
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
			"",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			template := MustNew(test.Options, test.Use...).MustParseString(test.Template)
			var buf bytes.Buffer
			n, err := template.Exec(&buf, test.ContextRun1)
			if test.ExpectError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err, test.ExpectError)
			}
			require.Equal(t, test.ExpectOutputRun1, buf.String())
			require.Equal(t, test.ExpectContextAfterRun1, test.ContextRun1)
			require.Len(t, test.ExpectOutputRun1, n)

			// run again with the second context
			buf.Reset()
			n, err = template.Exec(&buf, test.ContextRun2)
			if test.ExpectError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err, test.ExpectError)
			}
			require.Equal(t, test.ExpectOutputRun2, buf.String())
			require.Equal(t, test.ExpectContextAfterRun2, test.ContextRun2)
			require.Len(t, test.ExpectOutputRun2, n)
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
			var buf bytes.Buffer
			templates[templateIndex].MustExec(&buf, msg)
			require.Equal(t, fmt.Sprintf("Hello %d %d", templateIndex, i), buf.String())
		}
	}

	for i := 0; i < 1000; i++ {
		t.Run(strconv.Itoa(i), run(i))
	}
}

func TestFmtSprintf(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$fmt.Printf(fmt.Sprintf("Hello %s", "World"))$>`)
	var buf bytes.Buffer
	template.MustExec(&buf, nil)
	require.Equal(t, "Hello World", buf.String())
}

func TestPanic(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$panic("Oh no")$>`)
	var buf bytes.Buffer
	_, err := template.Exec(&buf, nil)
	require.EqualError(t, err, "Oh no")
}

func TestNoStartOrEnd(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.StartTokens = []rune{}
	template.EndTokens = []rune{}
	template.MustParseString(`fmt.Printf(fmt.Sprintf("Hello %s", "World"))`)
	var buf bytes.Buffer
	template.MustExec(&buf, nil)
	require.Equal(t, "Hello World", buf.String())
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

		var buf bytes.Buffer
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())

		buf.Reset()
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())
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

		var buf bytes.Buffer
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())

		buf.Reset()
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())
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

		var buf bytes.Buffer
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())

		buf.Reset()
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())
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

		var buf bytes.Buffer
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())

		buf.Reset()
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())
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

		var buf bytes.Buffer
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())

		buf.Reset()
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())
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

		var buf bytes.Buffer
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())

		buf.Reset()
		template.MustExec(&buf, nil)
		require.Equal(t, "Hello World", buf.String())
	})
}

func TestImplicitReturn(t *testing.T) {
	tests := []struct {
		key    string
		value  interface{}
		expect string
	}{
		{
			"Bool",
			false,
			"false",
		},
		{
			"Int",
			1,
			"1",
		},
		{
			"Uint",
			uint(1),
			"1",
		},
		{
			"Float",
			1.2,
			"1.2",
		},
		{
			"String",
			"Foo",
			"Foo",
		},
		{
			"Func",
			func() {},
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			template := MustNew(interp.Options{}, stdlib.Symbols)
			template.StartTokens = nil
			template.EndTokens = nil
			template.MustParseString(fmt.Sprintf(`context["%s"]`, test.key))
			var buf bytes.Buffer
			template.MustExec(&buf, map[string]interface{}{test.key: test.value})
			require.Equal(t, test.expect, buf.String())
		})
	}
}
