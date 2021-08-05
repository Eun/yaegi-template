package yaegi_template

import (
	"io"
	"reflect"
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
		Imports      []Import
		Template     string
		ExpectOutput string
		ExpectError  string
	}{
		{
			"Hello Yaegi",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			nil,
			`<html><$print("Hello Yaegi")$></html>`,
			`<html>Hello Yaegi</html>`,
			"",
		},
		{
			"Error",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			nil,
			`<$ Hello $>`,
			"",
			`1:29: undefined: Hello`,
		},
		{
			"Import",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			nil,
			`<$import "net/url"$><$print(url.PathEscape("Hello World"))$>`,
			"Hello%20World",
			"",
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			template := MustNew(test.Options, test.Use...).MustImport(test.Imports...).MustParseString(test.Template)
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
		Imports                []Import
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
			[]Import{{Name: "", Path: "fmt"}},
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
			[]Import{{Name: "", Path: "fmt"}},
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
			[]Import{{Name: "", Path: "fmt"}},
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
			[]Import{{Name: "", Path: "fmt"}},
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
			[]Import{{Name: "", Path: "fmt"}},
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
			template := MustNew(test.Options, test.Use...).MustImport(test.Imports...).MustParseString(test.Template)
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
		MustParse(bytes.NewBufferString(`Hello <$ print("Yaegi") $>`))

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
		templates = append(templates,
			MustNew(interp.Options{}, stdlib.Symbols).
				MustImport(Import{Name: "", Path: "fmt"}).
				MustParseString(`<$fmt.Printf("Hello `+strconv.Itoa(i)+` %s", context.Message)$>`),
		)
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
	template := MustNew(interp.Options{}, stdlib.Symbols).
		MustImport(Import{Name: "", Path: "fmt"}).
		MustParseString(`<$fmt.Printf(fmt.Sprintf("Hello %s", "World"))$>`)
	var buf bytes.Buffer
	template.MustExec(&buf, nil)
	require.Equal(t, "Hello World", buf.String())
}

func TestPanic(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols).
		MustParseString(`<$panic("Oh no")$>`)
	var buf bytes.Buffer
	_, err := template.Exec(&buf, nil)
	require.EqualError(t, err, "Oh no")
}

func TestNoStartOrEnd(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols).
		MustImport(Import{Name: "", Path: "fmt"})
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
		}, stdlib.Symbols).
			MustImport(Import{Name: "", Path: "fmt"})
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
		}, stdlib.Symbols).
			MustImport(Import{Name: "", Path: "fmt"})
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
		}, stdlib.Symbols).
			MustImport(Import{Name: "", Path: "fmt"})
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

func TestTemplate_LazyParse(t *testing.T) {
	buf := bytes.NewReader([]byte(`Hello <$ print("World") $>`))
	MustNew(DefaultOptions(), DefaultSymbols()...).MustLazyParse(buf)
	pos, err := buf.Seek(0, io.SeekCurrent)
	require.NoError(t, err)
	require.Equal(t, int64(0), pos)
}

func TestTemplate_ExecWithoutParse(t *testing.T) {
	_, err := MustNew(DefaultOptions(), DefaultSymbols()...).Exec(nil, nil)
	require.EqualError(t, err, "template was never parsed")
}

func TestTemplate_ExecToNilWriter(t *testing.T) {
	buf := bytes.NewReader([]byte(`Hello <$ print("World") $>`))
	MustNew(DefaultOptions(), DefaultSymbols()...).MustLazyParse(buf).MustExec(nil, nil)
}

func TestTemplate_Import(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		tm := MustNew(DefaultOptions(), DefaultSymbols()...).
			MustParseString(`Hello <$ print(http.StatusOK) $>`).
			MustImport(Import{
				Path: "net/http",
			})
		var buf bytes.Buffer
		tm.MustExec(&buf, nil)
		require.Equal(t, "Hello 200", buf.String())
	})
	t.Run("double import", func(t *testing.T) {
		tm := MustNew(DefaultOptions(), DefaultSymbols()...).
			MustParseString(`Hello <$ print(http.StatusOK) $>`).
			MustImport(Import{
				Path: "net/http",
			}).
			MustImport(Import{
				Path: "net/http",
			})
		var buf bytes.Buffer
		tm.MustExec(&buf, nil)
		require.Equal(t, "Hello 200", buf.String())
	})
	t.Run("alias import", func(t *testing.T) {
		tm := MustNew(DefaultOptions(), DefaultSymbols()...).
			MustParseString(`Hello <$ print(h.StatusOK) $>`).
			MustImport(Import{
				Name: "h",
				Path: "net/http",
			}).
			MustImport(Import{
				Path: "net/http",
			})
		var buf bytes.Buffer
		tm.MustExec(&buf, nil)
		require.Equal(t, "Hello 200", buf.String())
	})
	t.Run("import before parse", func(t *testing.T) {
		tm := MustNew(DefaultOptions(), DefaultSymbols()...).
			MustImport(Import{
				Path: "net/http",
			}).
			MustParseString(`Hello <$ print(http.StatusOK) $>`)
		var buf bytes.Buffer
		tm.MustExec(&buf, nil)
		require.Equal(t, "Hello 200", buf.String())
	})
}

func TestTemplateWithAdditionalSymbols(t *testing.T) {
	t.Run("using New+Import", func(t *testing.T) {
		t.Run("separate namespace", func(t *testing.T) {
			var buf bytes.Buffer
			MustNew(
				DefaultOptions(),
				append(DefaultSymbols(), interp.Exports{
					"ext/ext": map[string]reflect.Value{
						"Foo": reflect.ValueOf(func() string {
							return "foo"
						}),
					},
				})...).
				MustImport(Import{
					Path: "ext",
				}).
				MustParseString(`Hello <$ print(ext.Foo()) $>`).
				MustExec(&buf, nil)
			require.Equal(t, "Hello foo", buf.String())
		})
		t.Run("in own namespace (dot import)", func(t *testing.T) {
			var buf bytes.Buffer
			MustNew(
				DefaultOptions(),
				append(DefaultSymbols(), interp.Exports{
					"ext/ext": map[string]reflect.Value{
						"Foo": reflect.ValueOf(func() string {
							return "foo"
						}),
					},
				})...).
				MustImport(Import{
					Name: ".",
					Path: "ext",
				}).
				MustParseString(`Hello <$ print(Foo()) $>`).
				MustExec(&buf, nil)
			require.Equal(t, "Hello foo", buf.String())
		})
		t.Run("in own namespace (private dot import)", func(t *testing.T) {
			var buf bytes.Buffer
			MustNew(
				DefaultOptions(),
				append(DefaultSymbols(), interp.Exports{
					"ext/ext": map[string]reflect.Value{
						"foo": reflect.ValueOf(func() string {
							return "foo"
						}),
					},
				})...).
				MustImport(Import{
					Name: ".",
					Path: "ext",
				}).
				MustParseString(`Hello <$ print(foo()) $>`).
				MustExec(&buf, nil)
			require.Equal(t, "Hello foo", buf.String())
		})
	})

	t.Run("Use() func", func(t *testing.T) {
		t.Run("separate namespace", func(t *testing.T) {
			var buf bytes.Buffer
			MustNew(
				DefaultOptions(),
				DefaultSymbols()...).
				MustUse(interp.Exports{
					"ext/ext": map[string]reflect.Value{
						"Foo": reflect.ValueOf(func() string {
							return "foo"
						}),
					},
				}).
				MustImport(Import{Path: "ext"}).
				MustParseString(`Hello <$ print(ext.Foo()) $>`).
				MustExec(&buf, nil)
			require.Equal(t, "Hello foo", buf.String())
		})
		t.Run("in own namespace (dot import)", func(t *testing.T) {
			var buf bytes.Buffer
			MustNew(
				DefaultOptions(),
				DefaultSymbols()...).
				MustUse(interp.Exports{
					"ext/ext": map[string]reflect.Value{
						"Foo": reflect.ValueOf(func() string {
							return "foo"
						}),
					},
				}).
				MustImport(Import{Name: ".", Path: "ext"}).
				MustParseString(`Hello <$ print(Foo()) $>`).
				MustExec(&buf, nil)
			require.Equal(t, "Hello foo", buf.String())
		})
		t.Run("in own namespace (private dot import)", func(t *testing.T) {
			var buf bytes.Buffer
			MustNew(
				DefaultOptions(),
				DefaultSymbols()...).
				MustUse(interp.Exports{
					"ext/ext": map[string]reflect.Value{
						"foo": reflect.ValueOf(func() string {
							return "foo"
						}),
					},
				}).
				MustImport(Import{Name: ".", Path: "ext"}).
				MustParseString(`Hello <$ print(foo()) $>`).
				MustExec(&buf, nil)
			require.Equal(t, "Hello foo", buf.String())
		})
	})
}

func TestMultiParts(t *testing.T) {
	t.Run("single line", func(t *testing.T) {
		template := MustNew(interp.Options{}, stdlib.Symbols).
			MustParseString(`Hello <$ if context.Name == "" { $>Unknown<$ } else { print(context.Name) } $>`)

		type Context struct {
			Name string
		}

		var buf bytes.Buffer
		template.MustExec(&buf, Context{Name: "Joe"})
		require.Equal(t, "Hello Joe", buf.String())
		buf.Reset()
		template.MustExec(&buf, Context{Name: ""})
		require.Equal(t, "Hello Unknown", buf.String())
	})

	t.Run("multi line", func(t *testing.T) {
		template := MustNew(interp.Options{}, stdlib.Symbols).
			MustParseString(`Hello
<$- 
print(" ")
if context.Name == "" { -$>
	Unknown
<$- } else {
	print(context.Name)
} -$>`)

		type Context struct {
			Name string
		}

		var buf bytes.Buffer
		template.MustExec(&buf, Context{Name: "Joe"})
		require.Equal(t, "Hello Joe", buf.String())
		buf.Reset()
		template.MustExec(&buf, Context{Name: ""})
		require.Equal(t, "Hello Unknown", buf.String())
	})
}
