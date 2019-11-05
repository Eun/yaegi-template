// Code generated by 'goexports debug/plan9obj'. DO NOT EDIT.

// +build go1.13,!go1.14

package stdlib

import (
	"debug/plan9obj"
	"reflect"
)

func init() {
	Symbols["debug/plan9obj"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Magic386":   reflect.ValueOf(plan9obj.Magic386),
		"Magic64":    reflect.ValueOf(plan9obj.Magic64),
		"MagicAMD64": reflect.ValueOf(plan9obj.MagicAMD64),
		"MagicARM":   reflect.ValueOf(plan9obj.MagicARM),
		"NewFile":    reflect.ValueOf(plan9obj.NewFile),
		"Open":       reflect.ValueOf(plan9obj.Open),

		// type definitions
		"File":          reflect.ValueOf((*plan9obj.File)(nil)),
		"FileHeader":    reflect.ValueOf((*plan9obj.FileHeader)(nil)),
		"Section":       reflect.ValueOf((*plan9obj.Section)(nil)),
		"SectionHeader": reflect.ValueOf((*plan9obj.SectionHeader)(nil)),
		"Sym":           reflect.ValueOf((*plan9obj.Sym)(nil)),
	}
}
