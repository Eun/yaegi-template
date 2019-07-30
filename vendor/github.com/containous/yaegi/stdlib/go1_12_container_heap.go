// +build go1.12,!go1.13

package stdlib

// Code generated by 'goexports container/heap'. DO NOT EDIT.

import (
	"container/heap"
	"reflect"
)

func init() {
	Symbols["container/heap"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Fix":    reflect.ValueOf(heap.Fix),
		"Init":   reflect.ValueOf(heap.Init),
		"Pop":    reflect.ValueOf(heap.Pop),
		"Push":   reflect.ValueOf(heap.Push),
		"Remove": reflect.ValueOf(heap.Remove),

		// type definitions
		"Interface": reflect.ValueOf((*heap.Interface)(nil)),

		// interface wrapper definitions
		"_Interface": reflect.ValueOf((*_container_heap_Interface)(nil)),
	}
}

// _container_heap_Interface is an interface wrapper for Interface type
type _container_heap_Interface struct {
	WLen  func() int
	WLess func(i int, j int) bool
	WPop  func() interface{}
	WPush func(x interface{})
	WSwap func(i int, j int)
}

func (W _container_heap_Interface) Len() int               { return W.WLen() }
func (W _container_heap_Interface) Less(i int, j int) bool { return W.WLess(i, j) }
func (W _container_heap_Interface) Pop() interface{}       { return W.WPop() }
func (W _container_heap_Interface) Push(x interface{})     { W.WPush(x) }
func (W _container_heap_Interface) Swap(i int, j int)      { W.WSwap(i, j) }
