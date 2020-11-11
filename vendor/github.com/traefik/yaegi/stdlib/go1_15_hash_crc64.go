// Code generated by 'yaegi extract hash/crc64'. DO NOT EDIT.

// +build go1.15

package stdlib

import (
	"go/constant"
	"go/token"
	"hash/crc64"
	"reflect"
)

func init() {
	Symbols["hash/crc64"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Checksum":  reflect.ValueOf(crc64.Checksum),
		"ECMA":      reflect.ValueOf(constant.MakeFromLiteral("14514072000185962306", token.INT, 0)),
		"ISO":       reflect.ValueOf(constant.MakeFromLiteral("15564440312192434176", token.INT, 0)),
		"MakeTable": reflect.ValueOf(crc64.MakeTable),
		"New":       reflect.ValueOf(crc64.New),
		"Size":      reflect.ValueOf(constant.MakeFromLiteral("8", token.INT, 0)),
		"Update":    reflect.ValueOf(crc64.Update),

		// type definitions
		"Table": reflect.ValueOf((*crc64.Table)(nil)),
	}
}
