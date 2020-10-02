package yaegi_template

import "io"

type eofWriter struct{}

func (eofWriter) Write([]byte) (int, error) {
	return 0, io.EOF
}
