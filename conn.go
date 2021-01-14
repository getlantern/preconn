// Package preconn provides an implementation of net.Conn that allows insertion
// of data before the beginning of the underlying connection.
package preconn

import (
	"bytes"
	"errors"
	"io"
	"net"
)

// Conn is a net.Conn that supports replaying. Reads are not concurrency-safe.
type Conn struct {
	net.Conn
	head         io.Reader
	consumedHead bool
}

// Wrap wraps the supplied conn and inserting the given bytes at the head of the
// stream.
func Wrap(conn net.Conn, head []byte) *Conn {
	return WrapReader(conn, bytes.NewReader(head))
}

// WrapReader wraps the supplied conn, reading from 'head' first.
func WrapReader(conn net.Conn, head io.Reader) *Conn {
	return &Conn{conn, head, false}
}

// Read implements the method from net.Conn and first consumes the head before
// using the underlying connection.
func (conn *Conn) Read(b []byte) (n int, err error) {
	if conn.consumedHead {
		return conn.Conn.Read(b)
	}
	// "Read conventionally returns what is available instead of waiting for more."
	// - https://golang.org/pkg/io/#Reader
	// Thus if we read off conn.head, we return immediately rather than continuing to conn.Conn.
	n, err = conn.head.Read(b)
	if errors.Is(err, io.EOF) {
		err = nil
		conn.consumedHead = true
		if n == 0 {
			return conn.Read(b)
		}
	}
	return
}
