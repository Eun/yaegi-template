// Code generated by 'yaegi extract crypto/subtle'. DO NOT EDIT.

// +build go1.15,!go1.16

package stdlib

import (
	"crypto/subtle"
	"reflect"
)

func init() {
	Symbols["crypto/subtle"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"ConstantTimeByteEq":   reflect.ValueOf(subtle.ConstantTimeByteEq),
		"ConstantTimeCompare":  reflect.ValueOf(subtle.ConstantTimeCompare),
		"ConstantTimeCopy":     reflect.ValueOf(subtle.ConstantTimeCopy),
		"ConstantTimeEq":       reflect.ValueOf(subtle.ConstantTimeEq),
		"ConstantTimeLessOrEq": reflect.ValueOf(subtle.ConstantTimeLessOrEq),
		"ConstantTimeSelect":   reflect.ValueOf(subtle.ConstantTimeSelect),
	}
}
