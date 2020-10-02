package yaegi_template

import (
	"bytes"

	"go.uber.org/atomic"
)

type outputBuffer struct {
	buf           *bytes.Buffer
	discardWrites *atomic.Bool
}

func newOutputBuffer(discardWrites bool) *outputBuffer {
	return &outputBuffer{
		buf:           bytes.NewBuffer(nil),
		discardWrites: atomic.NewBool(discardWrites),
	}
}

func (ob *outputBuffer) Write(p []byte) (int, error) {
	if ob.discardWrites.Load() {
		return len(p), nil
	}
	return ob.buf.Write(p)
}

func (ob *outputBuffer) Reset() {
	ob.buf.Reset()
}

func (ob *outputBuffer) Bytes() []byte {
	return ob.buf.Bytes()
}

func (ob *outputBuffer) DiscardWrites(v bool) {
	ob.discardWrites.Store(v)
}
