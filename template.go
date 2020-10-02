package yaegi_template

import (
	"io"
	"strings"

	"reflect"

	"bytes"

	"bufio"

	"unicode/utf8"

	"errors"

	"fmt"

	"io/ioutil"
	"sync"

	"go/parser"
	"go/scanner"
	"go/token"

	"go/ast"

	"github.com/traefik/yaegi/interp"
)

type Template struct {
	options        interp.Options
	use            []interp.Exports
	templateReader io.Reader
	consumedReader bool
	imports        importSymbols
	StartTokens    []rune
	EndTokens      []rune
	interp         *interp.Interpreter
	outputBuffer   *outputBuffer
	mu             sync.Mutex
}

func New(options interp.Options, use ...interp.Exports) (*Template, error) {
	t := &Template{
		options:     options,
		use:         make([]interp.Exports, len(use)),
		StartTokens: []rune("<$"),
		EndTokens:   []rune("$>"),
	}

	// copy use so we can be sure not to modify them
	for i := range use {
		t.use[i] = make(interp.Exports)
		for packageName, funcMap := range use[i] {
			t.use[i][packageName] = make(map[string]reflect.Value)
			for funcName, funcReference := range funcMap {
				t.use[i][packageName][funcName] = funcReference
			}
		}
	}
	return t, nil
}

func MustNew(options interp.Options, use ...interp.Exports) *Template {
	t, err := New(options, use...)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) Parse(reader io.Reader) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	// maybe in the future we parse the template here
	// for now we don't
	t.templateReader = reader

	t.outputBuffer = newOutputBuffer(true)
	t.options.Stdout = t.outputBuffer

	t.interp = interp.New(t.options)

	for i := 0; i < len(t.use); i++ {
		t.interp.Use(t.use[i])
	}

	// import fmt
	return t.importSymbol(importSymbol{
		Name: "",
		Path: "fmt",
	})
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
	t.mu.Lock()
	defer t.mu.Unlock()
	// to execute a template multiple times we must be able to seek the reader back
	// to the start, if we cannot seek back we fail
	if t.consumedReader {
		seek, ok := t.templateReader.(io.Seeker)
		if !ok {
			return 0, errors.New("unable to consume template reader again")
		}
		if _, err := seek.Seek(0, io.SeekStart); err != nil {
			return 0, fmt.Errorf("unable to seek: %w", err)
		}
	}
	t.consumedReader = true

	if len(t.StartTokens) == 0 || len(t.EndTokens) == 0 {
		code, err := ioutil.ReadAll(t.templateReader)
		if err != nil {
			return 0, fmt.Errorf("unable to read template reader: %w", err)
		}
		return t.execCode(string(code), writer, context)
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

		n, werr = t.execCode(codeBuffer.String(), writer, context)
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
func (t *Template) execCode(code string, out io.Writer, context interface{}) (int, error) {
	if err := t.evalImports(&code); err != nil {
		return 0, err
	}
	if context != nil {
		// do we need to
		t.interp.Use(interp.Exports{
			"internal": map[string]reflect.Value{
				"context": reflect.ValueOf(context),
			},
		})

		// always reimport internal
		if err := t.safeEval(`import . "internal"`); err != nil {
			return 0, err
		}
	}

	// make sure the buffer is empty
	t.outputBuffer.DiscardWrites(false)
	if err := t.safeEval(code); err != nil {
		return 0, err
	}
	n, err := out.Write(t.outputBuffer.Bytes())
	t.outputBuffer.DiscardWrites(true)
	t.outputBuffer.Reset()
	return n, err
}

func (t *Template) safeEval(code string) (err error) {
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

// evalImports finds all "import" lines evaluates them and removes them from the code
func (t *Template) evalImports(code *string) error {
	var ok bool
	ok, err := t.hasPackage(*code)
	if err != nil {
		return err
	}
	var c string
	if !ok {
		c = "package main\n" + *code
	} else {
		c = *code
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", c, parser.ImportsOnly)
	if err != nil {
		return err
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.IMPORT {
			continue
		}

		syms := make(importSymbols, 0, len(genDecl.Specs))
		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}

			sym := importSymbol{
				Name: "",
				Path: strings.TrimFunc(importSpec.Path.Value, func(r rune) bool {
					return r == '`' || r == '"'
				}),
			}

			if importSpec.Name != nil {
				sym.Name = importSpec.Name.Name
			}

			syms = append(syms, sym)
		}

		if err = t.importSymbol(syms...); err != nil {
			return err
		}

		pos := int(genDecl.Pos()) - 1
		end := int(genDecl.End()) - 1
		c = c[:pos] + strings.Repeat(" ", end-pos) + c[end:]
	}

	// remove package main\n
	*code = c[13:]

	return nil
}

// hasPackage returns true when the code has a 'package' line
func (*Template) hasPackage(s string) (bool, error) {
	_, err := parser.ParseFile(token.NewFileSet(), "", s, parser.PackageClauseOnly)
	if err != nil {
		errList, ok := err.(scanner.ErrorList)
		if !ok {
			return false, err
		}
		if len(errList) == 0 {
			return false, err
		}
		if !strings.HasPrefix(errList[0].Msg, fmt.Sprintf("expected '%s', found", token.PACKAGE)) {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (t *Template) importSymbol(imports ...importSymbol) error {
	var symbolsToImport importSymbols
	for _, symbol := range imports {
		if !t.imports.Contains(symbol) {
			symbolsToImport = append(symbolsToImport, symbol)
		}
	}

	if len(symbolsToImport) == 0 {
		return nil
	}

	if err := t.safeEval(symbolsToImport.ImportBlock()); err != nil {
		return err
	}
	t.imports = append(t.imports, symbolsToImport...)
	return nil
}
