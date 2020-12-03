package yaegi_template

import (
	"io"
	"os"
	"strings"

	"github.com/traefik/yaegi/stdlib"

	"reflect"

	"bytes"

	"fmt"

	"sync"

	"go/parser"
	"go/scanner"
	"go/token"

	"go/ast"

	"github.com/traefik/yaegi/interp"

	"github.com/Eun/yaegi-template/codebuffer"
)

type Template struct {
	options        interp.Options
	use            []interp.Exports
	templateReader io.Reader
	imports        importSymbols
	StartTokens    []rune
	EndTokens      []rune
	interp         *interp.Interpreter
	outputBuffer   *outputBuffer
	codeBuffer     *codebuffer.CodeBuffer
	mu             sync.Mutex
}

// DefaultOptions return the default options for the New and MustNew functions.
func DefaultOptions() interp.Options {
	return interp.Options{
		GoPath:    os.Getenv("GOPATH"),
		BuildTags: nil,
		Stdin:     nil,
		Stdout:    nil,
		Stderr:    nil,
	}
}

// DefaultImports return the default imports for the New and MustNew functions.
func DefaultImports() []interp.Exports {
	return []interp.Exports{stdlib.Symbols}
}

func New(
	options interp.Options, //nolint:gocritic // disable hugeParam: options is heavy
	use ...interp.Exports) (*Template, error) {
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

func MustNew(
	options interp.Options, //nolint:gocritic // disable hugeParam: options is heavy
	use ...interp.Exports) *Template {
	t, err := New(options, use...)
	if err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) Parse(reader io.Reader) error {
	if err := t.LazyParse(reader); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	it, err := t.codeBuffer.Iterator()
	if err != nil {
		return err
	}

	// parse everything now
	for it.Next() {
	}
	return it.Error()
}

func (t *Template) MustParse(r io.Reader) *Template {
	if err := t.Parse(r); err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) LazyParse(reader io.Reader) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	// maybe in the future we parse the template here
	// for now we don't
	t.templateReader = reader

	t.outputBuffer = newOutputBuffer(true)
	t.codeBuffer = codebuffer.New(reader, t.StartTokens, t.EndTokens)
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

func (t *Template) MustLazyParse(r io.Reader) *Template {
	if err := t.LazyParse(r); err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) ParseString(s string) error {
	return t.Parse(bytes.NewReader([]byte(s)))
}

func (t *Template) ParseBytes(b []byte) error {
	return t.Parse(bytes.NewReader(b))
}

func (t *Template) MustParseString(s string) *Template {
	if err := t.ParseString(s); err != nil {
		panic(err.Error())
	}
	return t
}

func (t *Template) MustParseBytes(b []byte) error {
	if err := t.MustParseBytes(b); err != nil {
		panic(err.Error())
	}
	return nil
}

func (t *Template) Exec(writer io.Writer, context interface{}) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	it, err := t.codeBuffer.Iterator()
	if err != nil {
		return 0, err
	}

	total := 0

	for it.Next() {
		part := it.Value()
		switch part.Type {
		case codebuffer.CodePartType:
			n, err := t.execCode(string(part.Content), writer, context)
			if err != nil {
				return total, err
			}
			if n > 0 {
				total += n
			}
		case codebuffer.TextPartType:
			n, err := writer.Write(part.Content)
			if err != nil {
				return total, err
			}
			if n > 0 {
				total += n
			}
		}
	}

	return total, it.Error()
}

func (t *Template) MustExec(writer io.Writer, context interface{}) {
	if _, err := t.Exec(writer, context); err != nil {
		panic(err.Error())
	}
}

type RuneReader interface {
	ReadRune() (rune, int, error)
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
		if _, err := t.safeEval(`import . "internal"`); err != nil {
			return 0, err
		}
	}

	// make sure the buffer is empty
	t.outputBuffer.DiscardWrites(false)
	res, err := t.safeEval(code)
	if err != nil {
		return 0, err
	}

	if t.outputBuffer.Length() == 0 {
		// implicit write
		fmt.Fprint(t.outputBuffer, printValue(res))
	}
	n, err := out.Write(t.outputBuffer.Bytes())
	t.outputBuffer.DiscardWrites(true)
	t.outputBuffer.Reset()
	return n, err
}

func (t *Template) safeEval(code string) (res reflect.Value, err error) {
	if strings.TrimSpace(code) == "" {
		return reflect.Value{}, nil
	}

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

	res, err = t.interp.Eval(code)
	if err != nil {
		return res, err
	}
	return res, err
}

func printValue(v reflect.Value) string {
	if !v.IsValid() || !v.CanInterface() {
		return ""
	}

	switch x := v.Interface().(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(x)
	case string:
		return x
	default:
		return ""
	}
}

// evalImports finds all "import" lines evaluates them and removes them from the code.
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

		if err := t.importSymbol(syms...); err != nil {
			return err
		}

		pos := int(genDecl.Pos()) - 1
		end := int(genDecl.End()) - 1
		c = c[:pos] + strings.Repeat(" ", end-pos) + c[end:]
	}

	// remove package main\n
	*code = c[f.Name.End():]

	return nil
}

// hasPackage returns true when the code has a 'package' line.
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

	if _, err := t.safeEval(symbolsToImport.ImportBlock()); err != nil {
		return err
	}
	t.imports = append(t.imports, symbolsToImport...)
	return nil
}
