package yaegi_template

import (
	"io"
	"testing"

	"bytes"

	"bufio"

	"errors"

	"reflect"

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
			`<html><$fmt.Fprint(out, "Hello Yaegi")$></html>`,
			`<html>Hello Yaegi</html>`,
			nil,
		},
		{
			"Func",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<html><$func Foo(text string) {
	fmt.Fprintf(out, "Hello %s", text)
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
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.Name, func(t *testing.T) {
			template := MustNew(test.Options, test.Use...).MustParseString(test.Template)
			var buf bytes.Buffer
			if _, err := template.Exec(&buf, nil); !equalError(test.ExpectError, err) {
				t.Fatalf("expected %#v, got %#v", test.ExpectError, err)
			}
			if test.ExpectBuffer != buf.String() {
				t.Fatalf("expected %#v, got %#v", test.ExpectBuffer, buf.String())
			}
		})
	}
}

func TestExecWithContext(t *testing.T) {
	type MessageContext struct {
		Message string
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
			"Hello Yaegi",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$fmt.Fprint(out, context["Message"])$>`,
			MessageContext{"Hello Yaegi"},
			MessageContext{"Hello Yaegi"},
			`Hello Yaegi`,
			nil,
		},
		{
			"Hello Yaegi (ptr)",
			interp.Options{},
			[]interp.Exports{stdlib.Symbols},
			`<$fmt.Fprint(out, context["Message"])$>`,
			&MessageContext{"Hello Yaegi"},
			&MessageContext{"Hello Yaegi"},
			`Hello Yaegi`,
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
		MustParse(bytes.NewBufferString(`Hello <$ fmt.Fprintf(out, "Yaegi") $>`))

	var buf bytes.Buffer
	if _, err := template.Exec(&buf, nil); err != nil {
		t.Fatal("exec failed")
	}
	if "Hello Yaegi" != buf.String() {
		t.Fatalf(`expected "Hello Yaegi", got %#v`, buf.String())
	}
}

func TestDoubleExec(t *testing.T) {
	type MessageContext struct {
		Message string
	}
	template := MustNew(interp.Options{}, stdlib.Symbols)
	template.MustParseString(`<$fmt.Fprintf(out, "Hello %s", context["Message"])$>`)

	var buf1 bytes.Buffer
	template.MustExec(&buf1, MessageContext{Message: "Yaegi"})
	if "Hello Yaegi" != buf1.String() {
		t.Fatalf(`expected "Hello Yaegi", got %#v`, buf1.String())
	}

	var buf2 bytes.Buffer
	template.MustExec(&buf2, MessageContext{Message: "World"})

	if "Hello World" != buf2.String() {
		t.Fatalf(`expected "Hello World", got %#v`, buf2.String())
	}
}
