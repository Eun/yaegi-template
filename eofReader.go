package yaegi_template

import "io"

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) {
	return 0, io.EOF
}
