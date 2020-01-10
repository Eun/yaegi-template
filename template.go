package yaegi_template

import (
	"io"

	"reflect"

	"bytes"

	"bufio"

	"unicode/utf8"

	"errors"

	"fmt"

	"github.com/containous/yaegi/interp"
	"io/ioutil"
)

type Template struct {
	options        interp.Options
	use            []interp.Exports
	templateReader io.Reader
	consumedReader bool
	StartTokens    []rune
	EndTokens      []rune
	interp         *interp.Interpreter
	outputBuffer   *bytes.Buffer
}

func New(options interp.Options, use ...interp.Exports) (*Template, error) {
	return &Template{
		options:     options,
		use:         use,
		StartTokens: []rune("<$"),
		EndTokens:   []rune("$>"),
	}, nil
}

func MustNew(options interp.Options, use ...interp.Exports) *Template {
	t, err := New(options, use...)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) Parse(reader io.Reader) error {
	// maybe in the future we parse the template here
	// for now we don't
	t.templateReader = reader

	t.interp = interp.New(t.options)

	t.outputBuffer = bytes.NewBuffer(nil)

	t.hijackOs()
	t.hijackFmt()

	for i := 0; i < len(t.use); i++ {
		t.interp.Use(t.use[i])
	}

	t.interp.Eval(`import "fmt"`)

	return nil
}

func (t *Template) MustParse(r io.Reader) *Template {
	if err := t.Parse(r); err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) ParseString(s string) error {
	return t.Parse(bytes.NewReader([]byte(s)))
}

func (t *Template) MustParseString(s string) *Template {
	if err := t.ParseString(s); err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) Exec(writer io.Writer, context interface{}) (int, error) {
	// to execute a template multiple times we must be able to seek the reader back
	// to the start, if we cannot seek back we fail
	if t.consumedReader {
		seek, ok := t.templateReader.(io.Seeker)
		if !ok {
			return 0, errors.New("unable to consume template reader again")
		}
		seek.Seek(0, io.SeekStart)
	}
	t.consumedReader = true



	if len(t.StartTokens) == 0 || len(t.EndTokens) == 0 {
		code, err := ioutil.ReadAll(t.templateReader)
		if err != nil {
			return 0, fmt.Errorf("unable to read tempalte reader: %w", err)
		}
		return t.runCode(string(code), writer, context)
	}

	r := bufio.NewReader(t.templateReader)


	total := 0
	for {
		n, rerr, werr := skipIdent(t.StartTokens, r, writer)
		total += n
		if rerr != nil {
			if rerr == io.EOF {
				return total, nil
			}
			return total, rerr
		}
		if werr != nil {
			return total, werr
		}

		var codeBuffer bytes.Buffer

		n, rerr, werr = skipIdent(t.EndTokens, r, &codeBuffer)
		if rerr != nil {
			if rerr == io.EOF {
				return total, nil
			}
			return total, rerr
		}
		if werr != nil {
			return total, werr
		}

		n, werr = t.runCode(codeBuffer.String(), writer, context)
		total += n
		if werr != nil {
			return total, werr
		}
	}
}

func (t *Template) MustExec(writer io.Writer, context interface{}) {
	if _, err := t.Exec(writer, context); err != nil {
		panic(err.Error())
	}
}

type RuneReader interface {
	ReadRune() (rune, int, error)
}

// This function needs refactoring
func skipIdent(token []rune, reader RuneReader, writer io.Writer) (int, error, error) {
	var buf bytes.Buffer
	i := 0
	total := 0
	size := len(token)
	for {
		r, _, rerr := reader.ReadRune()
	c:
		if r != 0 && r != utf8.RuneError {
			if r == token[i] {
				if i == size-1 {
					// we found the token
					i = 0
					return total, rerr, nil
				} else {
					buf.WriteRune(r)
					i++
				}
			} else {
				// not our token?
				// we have something in the buffer?
				// write it
				if i > 0 {
					n, werr := writer.Write(buf.Bytes())
					total += n
					if werr != nil {
						return total, rerr, werr
					}
					buf.Reset()
					i = 0
					goto c
				}

				// write the non matching rune
				i = 0
				n, werr := writer.Write([]byte(string([]rune{r})))
				total += n
				if werr != nil {
					return total, rerr, werr
				}
			}
		}

		// an error on reading
		if rerr != nil {
			if buf.Len() > 0 {
				n, werr := writer.Write(buf.Bytes())
				return total + n, rerr, werr
			}
			return total, rerr, nil
		}
	}
}
func (t *Template) runCode(code string, out io.Writer, context interface{}) (int, error) {
	if context != nil {
		t.interp.Use(interp.Exports{
			"internal": map[string]reflect.Value{
				"context": reflect.ValueOf(context),
			},
		})
		// reimport so we have the correct context values
		if _, err := t.interp.Eval(`import . "internal"`); err != nil {
			return 0, err
		}
	}

	if err := t.evalCode(code); err != nil {
		return 0, err
	}
	n, err := out.Write(t.outputBuffer.Bytes())
	t.outputBuffer.Reset()
	return n, err
}

func (t *Template) evalCode(code string) (err error) {
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		switch v := e.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("%v", v)
		}
	}()

	_, err = t.interp.Eval(code)
	if err != nil {
		return err
	}
	return err
}

func (t *Template) hijackOs() {
	for i, e := range t.use {
		if _, ok := e["os"]; ok {
			t.use[i]["os"]["Stdout"] = reflect.ValueOf(t.outputBuffer)
		}
	}
}

func (t *Template) hijackFmt() {
	print := func(a ...interface{}) (int, error) {
		return fmt.Fprint(t.outputBuffer, a...)
	}

	printf := func(format string, a ...interface{}) (int, error) {
		return fmt.Fprintf(t.outputBuffer, format, a...)
	}

	println := func(a ...interface{}) (int, error) {
		return fmt.Fprintln(t.outputBuffer, a...)
	}

	for i, e := range t.use {
		if _, ok := e["fmt"]; ok {
			t.use[i]["fmt"]["Print"] = reflect.ValueOf(print)
			t.use[i]["fmt"]["Printf"] = reflect.ValueOf(printf)
			t.use[i]["fmt"]["Println"] = reflect.ValueOf(println)
		}
	}
}
