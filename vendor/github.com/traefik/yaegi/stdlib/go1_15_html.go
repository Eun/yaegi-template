// Code generated by 'yaegi extract html'. DO NOT EDIT.

// +build go1.15

package stdlib

import (
	"html"
	"reflect"
)

func init() {
	Symbols["html"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"EscapeString":   reflect.ValueOf(html.EscapeString),
		"UnescapeString": reflect.ValueOf(html.UnescapeString),
	}
}
