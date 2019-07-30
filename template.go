package yaegi_template

import (
	"io"

	"reflect"

	"bytes"

	"bufio"

	"unicode/utf8"

	"errors"

	"github.com/Eun/go-convert"
	"github.com/containous/yaegi/interp"
)

var defaultContext = map[string]interface{}{}

type Template struct {
	options        interp.Options
	use            []interp.Exports
	templateReader io.Reader
	consumedReader bool
	StartTokens    []rune
	EndTokens      []rune
	context        reflect.Value
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
	for i := 0; i < len(t.use); i++ {
		t.interp.Use(t.use[i])
	}

	t.outputBuffer = bytes.NewBuffer(nil)
	a := reflect.ValueOf(defaultContext)
	t.context = reflect.New(a.Type()).Elem()
	t.interp.Use(interp.Exports{
		"internal": map[string]reflect.Value{
			"out":     reflect.ValueOf(t.outputBuffer),
			"context": t.context,
		},
	})

	t.interp.Eval(`import "fmt"`)
	t.interp.Eval(`import . "internal"`)

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
	// revert the reader if we can
	if t.consumedReader {
		seek, ok := t.templateReader.(io.Seeker)
		if !ok {
			return 0, errors.New("unable to consume template reader again")
		}
		seek.Seek(0, io.SeekStart)
	}
	t.consumedReader = true

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
		m, err := convert.Convert(context, defaultContext)
		if err != nil {
			return 0, err
		}
		t.context.Set(reflect.ValueOf(m.(map[string]interface{})))
	}
	_, err := t.interp.Eval(code)
	if err != nil {
		return 0, err
	}
	n, err := out.Write(t.outputBuffer.Bytes())
	t.outputBuffer.Reset()
	return n, err
}
