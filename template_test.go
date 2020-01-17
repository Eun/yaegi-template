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

	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
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
			if _, err := template.Exec(&buf, nil); !equalError(test.ExpectError, err) {
				t.Fatalf("expected %#v, got %#v", test.ExpectError, err.Error())
			}
			if test.ExpectBuffer != buf.String() {
				t.Fatalf("expected %#v, got %#v", test.ExpectBuffer, buf.String())
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
		Name                  string
		Options               interp.Options
		Use                   []interp.Exports
		Template              string
		Context               interface{}
		ExpectContextAfterRun interface{}
		ExpectBuffer          string
		ExpectError           error
	}{
		{
			"Struct",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`Hello <$fmt.Printf("%s (%d)", context.Name, context.ID)$>`,
			User{10, "Yaegi"},
			User{10, "Yaegi"},
			`Hello Yaegi (10)`,
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
			nil,
		},

		{
			"Map",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$fmt.Printf("%d %s", context["Foo"], context["Bar"])$>`,
			map[string]interface{}{"Foo": 10, "Bar": "Joe"},
			map[string]interface{}{"Foo": 10, "Bar": "Joe"},
			`10 Joe`,
			nil,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			template := MustNew(test.Options, test.Use...).MustParseString(test.Template)
			var buf bytes.Buffer
			if _, err := template.Exec(&buf, test.Context); !equalError(test.ExpectError, err) {
				t.Fatalf("expected %#v, got %#v", test.ExpectError, err)
			}
			if test.ExpectBuffer != buf.String() {
				t.Fatalf("expected %#v, got %#v", test.ExpectBuffer, buf.String())
			}
			if !reflect.DeepEqual(test.ExpectContextAfterRun, test.Context) {
				t.Fatalf("expected %#v, got %#v", test.ExpectContextAfterRun, test.Context)
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
	if "Hello Yaegi" != buf.String() {
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

	for i := 0; i < 1000; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			templateIndex := i % len(templates)
			msg := MessageContext{Message: strconv.Itoa(i)}
			expect := fmt.Sprintf("Hello %d %d", templateIndex, i)
			var buf bytes.Buffer
			templates[templateIndex].MustExec(&buf, msg)
			if expect != buf.String() {
				t.Fatalf(`expected %s, got %#v`, expect, buf.String())
			}
		})
	}
}

func TestFmtSprintf(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$fmt.Printf(fmt.Sprintf("Hello %s", "World"))$>`)
	var buf1 bytes.Buffer
	template.MustExec(&buf1, nil)
	if "Hello World" != buf1.String() {
		t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
	}
}

func Test1(t *testing.T) {
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$fmt.Printf("Hello World")$>`)
	var buf1 bytes.Buffer
	template.MustExec(&buf1, nil)
	fmt.Println(buf1.String())
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
	if "Hello World" != buf1.String() {
		t.Fatalf(`expected "Hello World", got %#v`, buf1.String())
	}
}
