package server

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeoutConnZeroTimeout(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})
	assert.Same(t, left, NewTimeoutConn(left, 0))
}

func TestTimeoutConnReadTimeout(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	conn := NewTimeoutConn(left, 20*time.Millisecond)
	start := time.Now()
	_, err := conn.Read(make([]byte, 1))
	require.Error(t, err)

	var netErr net.Error
	require.ErrorAs(t, err, &netErr)
	assert.True(t, netErr.Timeout())
	assert.GreaterOrEqual(t, time.Since(start), 15*time.Millisecond)
}

func TestTimeoutConnWriteTimeout(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	conn := NewTimeoutConn(left, 20*time.Millisecond)
	start := time.Now()
	_, err := conn.Write([]byte("x"))
	require.Error(t, err)

	var netErr net.Error
	require.ErrorAs(t, err, &netErr)
	assert.True(t, netErr.Timeout())
	assert.GreaterOrEqual(t, time.Since(start), 15*time.Millisecond)
}

func TestTimeoutConnRefreshesReadDeadline(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	conn := NewTimeoutConn(left, 20*time.Millisecond)
	go func() {
		time.Sleep(10 * time.Millisecond)
		_, _ = io.WriteString(right, "a")
		time.Sleep(10 * time.Millisecond)
		_, _ = io.WriteString(right, "b")
	}()

	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, byte('a'), buf[0])

	_, err = conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, byte('b'), buf[0])
}

func TestTimeoutConnRefreshesWriteDeadline(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = left.Close()
		_ = right.Close()
	})

	conn := NewTimeoutConn(left, 20*time.Millisecond)
	go func() {
		time.Sleep(10 * time.Millisecond)
		buf := make([]byte, 1)
		_, _ = right.Read(buf)
		time.Sleep(10 * time.Millisecond)
		_, _ = right.Read(buf)
	}()

	_, err := conn.Write([]byte("a"))
	require.NoError(t, err)
	_, err = conn.Write([]byte("b"))
	require.NoError(t, err)
}

func TestTimeoutConnCloseUnblocksPendingRead(t *testing.T) {
	left, right := net.Pipe()
	t.Cleanup(func() {
		_ = right.Close()
	})

	conn := NewTimeoutConn(left, time.Second)
	errCh := make(chan error, 1)
	go func() {
		_, err := conn.Read(make([]byte, 1))
		errCh <- err
	}()

	time.Sleep(10 * time.Millisecond)
	require.NoError(t, conn.Close())

	err := <-errCh
	require.Error(t, err)
	assert.True(t, errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrClosedPipe))
}
