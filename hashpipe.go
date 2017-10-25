package hashpipe

import "io"
import "hash"

// NewWriter binds a hash.Hash to an io.Writer so that everything written to the resulting writer will also be hashed
func NewWriter(h hash.Hash) func(io.Writer) io.Writer {
	return func(w io.Writer) io.Writer {
		return io.MultiWriter(w, h)
	}
}

// NewReader binds a hash.Hash to an io.Reader so that everything written to the resulting writer will also be hashed
func NewReader(h hash.Hash) func(io.Reader) io.Reader {
	return func(r io.Reader) io.Reader {
		return io.TeeReader(r, h)
	}
}
