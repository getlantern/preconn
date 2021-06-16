package preconn

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	head = "hello "
	text = "world"
	full = "hello world"
)

func TestPreConn(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	if !assert.NoError(t, err) {
		return
	}
	defer l.Close()
	go func() {
		for {
			conn, err := l.Accept()
			if err == nil {
				conn.Write([]byte(text))
				conn.Close()
			}
		}
	}()

	conn, err := net.Dial("tcp", l.Addr().String())
	if !assert.NoError(t, err) {
		return
	}
	defer conn.Close()
	pconn := Wrap(conn, []byte(head))
	var buf bytes.Buffer
	b := make([]byte, 2)
	for {
		n, err := pconn.Read(b)
		if err == io.EOF {
			break
		}
		if !assert.NoError(t, err) {
			return
		}
		buf.Write(b[:n])
	}
	assert.Equal(t, full, buf.String(), "Read() multiple times should get the full data")

	conn, err = net.Dial("tcp", l.Addr().String())
	if !assert.NoError(t, err) {
		return
	}
	defer conn.Close()
	b = make([]byte, len(full))
	pconn = Wrap(conn, []byte(head))
	n, err := pconn.Read(b)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, head, string(b[:n]), "Read() should return first available data")
	n, err = pconn.Read(b[n:])
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, full, string(b))
}

// The intent of this test is to expose data races. Run with -race.
func TestConcurrentReads(t *testing.T) {
	const parallelism = 10

	a, b := net.Pipe()
	a = Wrap(a, []byte("hello from c2"))
	defer a.Close()
	defer b.Close()

	go io.Copy(b, rand.Reader)

	wg := new(sync.WaitGroup)
	errs := make(chan error, parallelism)
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 1024)
			_, err := a.Read(buf)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
}
