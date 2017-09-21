// Package preconn provides an implementation of net.Conn that allows insertion
// of data before the beginning of the underlying connection.
package preconn

import (
	"net"
)

// Conn is a net.Conn that supports replaying.
type Conn struct {
	net.Conn
	head []byte
}

// Wrap wraps the supplied conn and inserting the given bytes at the head of the
// stream.
func Wrap(conn net.Conn, head []byte) *Conn {
	return &Conn{
		Conn: conn,
		head: head,
	}
}

// Read implements the method from net.Conn and first consumes the head before
// using the underlying connection.
func (conn *Conn) Read(b []byte) (n int, err error) {
	nc := copy(b, conn.head)
	conn.head = conn.head[nc:]
	if nc >= len(b) {
		n = nc
	} else {
		n, err = conn.Conn.Read(b[nc:])
		n += nc
	}
	return
}
