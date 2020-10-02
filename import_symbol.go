package yaegi_template

import "strings"

type importSymbols []importSymbol

func (is importSymbols) Contains(symbol importSymbol) bool {
	for _, s := range is {
		if s.Equals(symbol) {
			return true
		}
	}
	return false
}

func (s importSymbols) ImportBlock() string {
	switch len(s) {
	case 0:
		return ""
	case 1:
		return "import " + s[0].ImportLine()
	default:
		var sb strings.Builder
		sb.WriteString("import (\n")
		for _, symbol := range s {
			sb.WriteString(symbol.ImportLine())
			sb.WriteRune('\n')
		}
		sb.WriteString(")")
		return sb.String()
	}
}

type importSymbol struct {
	Name string
	Path string
}

func (s importSymbol) Equals(symbol importSymbol) bool {
	return s.Name == symbol.Name && strings.EqualFold(s.Path, symbol.Path)
}

func (s importSymbol) ImportLine() string {
	if s.Name != "" {
		return s.Name + ` "` + s.Path + `"`
	}
	return `"` + s.Path + `"`
}
