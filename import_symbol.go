package yaegi_template

import (
	"fmt"
	"strings"
)

type importSymbols []Import

func (is importSymbols) Contains(symbol Import) bool {
	for _, s := range is {
		if s.Equals(symbol) {
			return true
		}
	}
	return false
}

func (is importSymbols) ImportBlock() string {
	switch len(is) {
	case 0:
		return ""
	case 1:
		return "import " + is[0].ImportLine()
	default:
		var sb strings.Builder
		sb.WriteString("import (\n")
		for _, symbol := range is {
			sb.WriteString(symbol.ImportLine())
			sb.WriteRune('\n')
		}
		sb.WriteString(")")
		return sb.String()
	}
}

type Import struct {
	Name string
	Path string
}

func (s Import) Equals(symbol Import) bool {
	return s.Name == symbol.Name && strings.EqualFold(s.Path, symbol.Path)
}

func (s Import) ImportLine() string {
	if s.Name != "" {
		return fmt.Sprintf("%s %q", s.Name, s.Path)
	}
	return fmt.Sprintf("%q", s.Path)
}
