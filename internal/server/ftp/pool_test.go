package ftp

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	ftplib "github.com/jlaffaye/ftp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoolConcurrentTransfers(t *testing.T) {
	for _, connections := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("connections=%d", connections), func(t *testing.T) {
			gate := make(chan struct{})
			var release sync.Once
			t.Cleanup(func() { release.Do(func() { close(gate) }) })
			started := make(chan struct{}, 16)
			var conns []*stubFTPConn
			var active atomic.Int32
			newSession := func() *Client {
				var busy atomic.Bool
				conn := &stubFTPConn{
					retrFunc: func(filename string) (io.ReadCloser, error) {
						assert.False(t, busy.Swap(true), "FTP session used concurrently")
						assert.LessOrEqual(t, active.Add(1), int32(connections))
						started <- struct{}{}
						<-gate
						return &stubReadCloser{
							data: bytes.NewBufferString(filename),
							closeFunc: func() {
								busy.Store(false)
								active.Add(-1)
							},
						}, nil
					},
				}
				conns = append(conns, conn)
				return &Client{cfg: (&config.ServerFTP{}).GetDefaults(), ftp: conn}
			}
			p := newPool(newSession())
			attempts := 0
			p.connectFunc = func() (*Client, error) {
				attempts++
				if len(conns) == connections {
					return nil, errors.New("connection limit reached")
				}
				return newSession(), nil
			}
			var workers sync.WaitGroup
			for i := range 3 {
				workers.Go(func() {
					for j := range 3 {
						filename := fmt.Sprintf("/file-%d-%d", i, j)
						var dest bytes.Buffer
						if err := p.Retrieve(filename, &dest); err != nil {
							t.Error(err)
							return
						}
						assert.Equal(t, filename, dest.String())
					}
				})
			}
			done := make(chan struct{})
			go func() {
				workers.Wait()
				close(done)
			}()
			t.Cleanup(func() {
				release.Do(func() { close(gate) })
				<-done
				assert.NoError(t, p.Close())
			})
			for range connections {
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					t.Fatal("FTP transfers did not overlap")
				}
			}
			if connections < 3 {
				require.Eventually(t, func() bool {
					p.mu.Lock()
					defer p.mu.Unlock()
					return p.limited
				}, 5*time.Second, time.Millisecond)
			}
			release.Do(func() { close(gate) })
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("FTP transfers did not finish")
			}
			assert.Len(t, conns, connections)
			assert.Equal(t, min(connections, 2), attempts)
			assert.NoError(t, p.Close())
			assert.NoError(t, p.Close())
			for _, conn := range conns {
				assert.Equal(t, 1, conn.quitCall)
			}
			require.ErrorIs(t, p.Retrieve("/after-close", io.Discard), net.ErrClosed)
			_, err := p.ReadDir("/")
			require.ErrorIs(t, err, net.ErrClosed)
		})
	}
}

func TestPoolRetiresFailedSession(t *testing.T) {
	p := newPool(&Client{cfg: (&config.ServerFTP{}).GetDefaults()})
	defer p.Close()
	failed, err := p.acquire()
	require.NoError(t, err)
	healthy := &Client{ftp: &stubFTPConn{}}
	p.connectFunc = func() (*Client, error) { return healthy, nil }
	other, err := p.acquire()
	require.NoError(t, err)
	failed.connectErr = errors.New("reconnect failed")
	p.release(failed)
	replacement := &Client{ftp: &stubFTPConn{}}
	p.connectFunc = func() (*Client, error) { return replacement, nil }
	next, err := p.acquire()
	require.NoError(t, err)
	assert.Same(t, replacement, next)
	p.release(next)
	p.release(other)
}

func TestPoolPreservesLastSessionReconnectFailure(t *testing.T) {
	conn := &stubFTPConn{retrErr: net.ErrClosed}
	dials := 0
	p := newPool(&Client{
		cfg: (&config.ServerFTP{}).GetDefaults(),
		ftp: conn,
		dial: func(string, netDialFunc, ...ftplib.DialOption) (ftpConn, error) {
			dials++
			return nil, errors.New("server unavailable")
		},
	})
	defer p.Close()
	require.ErrorIs(t, p.Retrieve("/first", io.Discard), net.ErrClosed)
	for range 3 {
		require.ErrorContains(t, p.Retrieve("/next", io.Discard), "server unavailable")
	}
	_, err := p.ReadDir("/source")
	require.ErrorContains(t, err, "server unavailable")
	assert.Equal(t, 1, dials, "a cached reconnect failure must fail fast for subsequent files")
	assert.Equal(t, 1, conn.quitCall)
}

func TestPoolCloseDuringRetrieve(t *testing.T) {
	started := make(chan struct{})
	finish := make(chan struct{})
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(finish) }) })
	conn := &stubFTPConn{
		retrFunc: func(string) (io.ReadCloser, error) {
			close(started)
			<-finish
			return io.NopCloser(bytes.NewBufferString("payload")), nil
		},
	}
	p := newPool(&Client{cfg: (&config.ServerFTP{}).GetDefaults(), ftp: conn})
	retrieved := make(chan error, 1)
	go func() { retrieved <- p.Retrieve("/file", io.Discard) }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("transfer did not start")
	}
	closed := make(chan error, 1)
	go func() { closed <- p.Close() }()
	select {
	case err := <-closed:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Close waited for an in-flight transfer")
	}
	assert.Zero(t, conn.quitCall, "a borrowed session must remain exclusive to its operation")
	release.Do(func() { close(finish) })
	select {
	case err := <-retrieved:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("transfer did not finish")
	}
	assert.Equal(t, 1, conn.quitCall, "a session returned after Close must be closed")
}

func TestPoolCloseWithPendingOperations(t *testing.T) {
	for _, connecting := range []bool{false, true} {
		t.Run(fmt.Sprintf("connecting=%t", connecting), func(t *testing.T) {
			first := &stubFTPConn{}
			second := &stubFTPConn{}
			p := newPool(&Client{cfg: (&config.ServerFTP{}).GetDefaults(), ftp: first})
			session, err := p.acquire()
			require.NoError(t, err)
			dialStarted := make(chan struct{})
			finishDial := make(chan struct{})
			if connecting {
				p.connectFunc = func() (*Client, error) {
					close(dialStarted)
					<-finishDial
					return &Client{ftp: second}, nil
				}
			} else {
				p.limited = true
			}
			acquired := make(chan error, 1)
			go func() {
				_, err := p.acquire()
				acquired <- err
			}()
			if connecting {
				select {
				case <-dialStarted:
				case <-time.After(5 * time.Second):
					t.Fatal("connection attempt did not start")
				}
			}
			closed := make(chan error, 1)
			go func() { closed <- p.Close() }()
			select {
			case err := <-closed:
				require.NoError(t, err)
			case <-time.After(5 * time.Second):
				t.Fatal("Close waited for a borrowed session or pending connection")
			}
			p.release(session)
			if connecting {
				close(finishDial)
			}
			select {
			case err := <-acquired:
				require.ErrorIs(t, err, net.ErrClosed)
			case <-time.After(5 * time.Second):
				t.Fatal("waiting operation did not stop after Close")
			}
			assert.Equal(t, 1, first.quitCall)
			if connecting {
				assert.Equal(t, 1, second.quitCall)
			}
		})
	}
}
