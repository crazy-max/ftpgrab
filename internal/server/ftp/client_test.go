package ftp

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/hashicorp/go-multierror"
	ftplib "github.com/jlaffaye/ftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type stubFTPConn struct {
	listEntries []*ftplib.Entry
	listErr     error
	listPath    string
	loginErr    error
	quitErr     error
	quitCall    int
	retrErr     error
	retrPath    string
	retrResp    io.ReadCloser
}

func (c *stubFTPConn) Login(_, _ string) error {
	return c.loginErr
}

func (c *stubFTPConn) List(path string) ([]*ftplib.Entry, error) {
	c.listPath = path
	return c.listEntries, c.listErr
}

func (c *stubFTPConn) Retr(path string) (io.ReadCloser, error) {
	c.retrPath = path
	return c.retrResp, c.retrErr
}

func (c *stubFTPConn) Quit() error {
	c.quitCall++
	return c.quitErr
}

type stubReadCloser struct {
	readErr  error
	closeErr error
	data     *bytes.Buffer
}

func (r *stubReadCloser) Read(p []byte) (int, error) {
	if r.readErr != nil {
		return 0, r.readErr
	}
	if r.data == nil {
		return 0, io.EOF
	}
	return r.data.Read(p)
}

func (r *stubReadCloser) Close() error {
	return r.closeErr
}

type timeoutError struct{}

func (timeoutError) Error() string {
	return "i/o timeout"
}

func (timeoutError) Timeout() bool {
	return true
}

func TestGetTLSMode(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		assert.Equal(t, tlsModeDisabled, getTLSMode(cfg))
	})

	t.Run("implicit", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		cfg.TLS = new(true)
		assert.Equal(t, tlsModeImplicit, getTLSMode(cfg))
	})

	t.Run("explicit", func(t *testing.T) {
		cfg := (&config.ServerFTP{}).GetDefaults()
		cfg.ExplicitTLS = new(true)
		assert.Equal(t, tlsModeExplicit, getTLSMode(cfg))
	})
}

func TestNewTimeoutDialFunc(t *testing.T) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = ln.Close()
	})

	accepted := make(chan net.Conn, 4)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				close(accepted)
				return
			}
			accepted <- conn
		}
	}()

	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	t.Run("disabled", func(t *testing.T) {
		dial := newTimeoutDialFunc(50*time.Millisecond, tlsModeDisabled, tlsConfig)
		conn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = conn.Close()
		})

		_, isTLS := conn.(*tls.Conn)
		assert.False(t, isTLS)

		serverConn := <-accepted
		t.Cleanup(func() {
			_ = serverConn.Close()
		})
	})

	t.Run("implicit", func(t *testing.T) {
		dial := newTimeoutDialFunc(50*time.Millisecond, tlsModeImplicit, tlsConfig)
		conn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = conn.Close()
		})

		_, isTLS := conn.(*tls.Conn)
		assert.True(t, isTLS)

		serverConn := <-accepted
		t.Cleanup(func() {
			_ = serverConn.Close()
		})
	})

	t.Run("explicit", func(t *testing.T) {
		dial := newTimeoutDialFunc(50*time.Millisecond, tlsModeExplicit, tlsConfig)

		controlConn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = controlConn.Close()
		})
		_, isTLS := controlConn.(*tls.Conn)
		assert.False(t, isTLS)

		serverControlConn := <-accepted
		t.Cleanup(func() {
			_ = serverControlConn.Close()
		})

		dataConn, err := dial("tcp", ln.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = dataConn.Close()
		})
		_, isTLS = dataConn.(*tls.Conn)
		assert.True(t, isTLS)

		serverDataConn := <-accepted
		t.Cleanup(func() {
			_ = serverDataConn.Close()
		})
	})
}

func TestClientConnect(t *testing.T) {
	t.Run("skips login without username", func(t *testing.T) {
		conn := &stubFTPConn{}
		client := &Client{
			addr: "ftp.example:21",
			dial: func(_ string, _ ...ftplib.DialOption) (ftpConn, error) {
				return conn, nil
			},
		}
		require.NoError(t, client.connect())
		assert.Equal(t, 0, conn.quitCall)
		assert.Equal(t, conn, client.ftp)
	})

	t.Run("closes connection when login fails", func(t *testing.T) {
		conn := &stubFTPConn{loginErr: errors.New("boom")}
		client := &Client{
			addr:     "ftp.example:21",
			username: "user",
			password: "secret",
			dial: func(_ string, _ ...ftplib.DialOption) (ftpConn, error) {
				return conn, nil
			},
		}
		err := client.connect()
		require.EqualError(t, err, "boom")
		assert.Equal(t, 1, conn.quitCall)
		assert.Nil(t, client.ftp)
	})

	t.Run("does not close connection on successful login", func(t *testing.T) {
		conn := &stubFTPConn{}
		client := &Client{
			addr:     "ftp.example:21",
			username: "user",
			password: "secret",
			dial: func(_ string, _ ...ftplib.DialOption) (ftpConn, error) {
				return conn, nil
			},
		}
		require.NoError(t, client.connect())
		assert.Equal(t, 0, conn.quitCall)
		assert.Equal(t, conn, client.ftp)
	})
}

func TestNewPathEnc(t *testing.T) {
	t.Run("normalizes utf8", func(t *testing.T) {
		pathenc, err := newPathEnc("utf-8")
		require.NoError(t, err)

		value, err := pathenc.Encode("/Проекты")
		require.NoError(t, err)
		assert.Equal(t, "/Проекты", value)
	})

	t.Run("supports windows1251", func(t *testing.T) {
		pathenc, err := newPathEnc("windows-1251")
		require.NoError(t, err)

		encoded, err := pathenc.Encode("/Проекты")
		require.NoError(t, err)
		assert.NotEqual(t, "/Проекты", encoded)

		decoded, err := pathenc.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, "/Проекты", decoded)
	})
}

func TestReadDirDecodesEntryNames(t *testing.T) {
	pathenc, err := newPathEnc("windows-1251")
	require.NoError(t, err)

	conn := &stubFTPConn{
		listEntries: []*ftplib.Entry{
			{
				Name: mustEncodeWindows1251(t, "Инструкция_по_заливке+договор"),
				Type: ftplib.EntryTypeFolder,
			},
		},
	}
	client := &Client{
		ftp:     conn,
		pathenc: pathenc,
	}

	entries, err := client.ReadDir("/Проекты")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "/\xcf\xf0\xee\xe5\xea\xf2\xfb", conn.listPath)
	assert.Equal(t, "Инструкция_по_заливке+договор", entries[0].Name())
	assert.True(t, entries[0].IsDir())
}

func TestReadDirLiteralPaths(t *testing.T) {
	for _, mlsd := range []bool{true, false} {
		for _, escape := range []bool{true, false} {
			t.Run(fmt.Sprintf("mlsd=%v/escape=%v", mlsd, escape), func(t *testing.T) {
				conn := newListingTestServer(t, listingServerOptions{mlsd: mlsd})
				client := &Client{
					cfg: &config.ServerFTP{EscapeRegexpMeta: new(escape)},
					ftp: serverConn{ServerConn: conn},
				}
				// Repeat with a relative source to check that CWD is restored.
				for _, dir := range []string{"/complete/Game.v1 [Collection]/café", "complete/Game.v1 [Collection]/café"} {
					entries, err := client.ReadDir(dir)
					require.NoError(t, err)
					require.Len(t, entries, 1)
					assert.Equal(t, "résumé.txt", entries[0].Name())
					cwd, err := conn.CurrentDir()
					require.NoError(t, err)
					assert.Equal(t, "/", cwd)
				}
			})
		}
	}
}

func TestReadDirRestoresDirectoryAfterListingFailure(t *testing.T) {
	conn := newListingTestServer(t, listingServerOptions{mlsd: true, failList: true})
	client := &Client{
		ftp: serverConn{ServerConn: conn},
	}
	_, err := client.ReadDir("/complete/Game.v1 [Collection]/café")
	require.Equal(t, &textproto.Error{Code: 550, Msg: "Listing failed"}, err)
	cwd, err := conn.CurrentDir()
	require.NoError(t, err)
	assert.Equal(t, "/", cwd)
}

func TestReadDirRestoresQuotedDirectory(t *testing.T) {
	conn := newListingTestServer(t, listingServerOptions{mlsd: true, home: `/home/a""b`})
	client := &Client{ftp: serverConn{ServerConn: conn}}
	for range 2 {
		_, err := client.ReadDir("complete/Game.v1 [Collection]/café")
		require.NoError(t, err)
	}
}

func TestReadDirReconnectsAfterRestoreFailure(t *testing.T) {
	for _, failList := range []bool{false, true} {
		t.Run(fmt.Sprintf("failList=%v", failList), func(t *testing.T) {
			conn := newListingTestServer(t, listingServerOptions{
				mlsd:        true,
				failList:    failList,
				failRestore: true,
			})
			next := &stubFTPConn{}
			client := &Client{
				ftp: serverConn{ServerConn: conn},
				dial: func(_ string, _ ...ftplib.DialOption) (ftpConn, error) {
					return next, nil
				},
			}
			_, err := client.ReadDir("complete/Game.v1 [Collection]/café")
			require.ErrorIs(t, err, errRestoreDir)
			if failList {
				var replyErr *textproto.Error
				require.ErrorAs(t, err, &replyErr)
				assert.Equal(t, 550, replyErr.Code)
				assert.Equal(t, "Listing failed", replyErr.Msg)
			}
			require.Nil(t, client.ftp)
			_, err = client.ReadDir("complete/other")
			require.NoError(t, err)
			assert.Equal(t, next, client.ftp)
			assert.Equal(t, "complete/other", next.listPath)
		})
	}
}

func TestRetrieveEncodesPath(t *testing.T) {
	pathenc, err := newPathEnc("windows-1251")
	require.NoError(t, err)

	conn := &stubFTPConn{retrErr: errors.New("stop")}
	client := &Client{
		ftp:     conn,
		pathenc: pathenc,
	}

	err = client.Retrieve("/Проекты/Инструкция.txt", nil)
	require.EqualError(t, err, "stop")
	assert.Equal(t, mustEncodeWindows1251(t, "/Проекты/Инструкция.txt"), conn.retrPath)
}

func TestReadDirReconnectsAfterTimeout(t *testing.T) {
	firstConn := &stubFTPConn{
		listErr: multierror.Append(errors.New("data stalled"), timeoutError{}),
	}
	secondConn := &stubFTPConn{
		listEntries: []*ftplib.Entry{
			{Name: "ok.txt", Type: ftplib.EntryTypeFile},
		},
	}

	dialCalls := 0
	client := &Client{
		ftp: firstConn,
		dial: func(_ string, _ ...ftplib.DialOption) (ftpConn, error) {
			dialCalls++
			if dialCalls == 1 {
				return secondConn, nil
			}
			return nil, errors.New("unexpected dial")
		},
	}

	_, err := client.ReadDir("/source")
	require.Error(t, err)
	assert.Equal(t, 1, firstConn.quitCall)
	assert.Nil(t, client.ftp)

	entries, err := client.ReadDir("/source")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "ok.txt", entries[0].Name())
	assert.Equal(t, 1, dialCalls)
	assert.Equal(t, secondConn, client.ftp)
}

func TestReadDirDoesNotRetryReconnectAfterFailedReconnect(t *testing.T) {
	firstConn := &stubFTPConn{
		listErr: multierror.Append(errors.New("data stalled"), timeoutError{}),
	}

	dialCalls := 0
	client := &Client{
		ftp: firstConn,
		dial: func(_ string, _ ...ftplib.DialOption) (ftpConn, error) {
			dialCalls++
			return nil, errors.New("tls: first record does not look like a TLS handshake")
		},
	}

	_, err := client.ReadDir("/source-a")
	require.Error(t, err)
	assert.Equal(t, 1, firstConn.quitCall)
	assert.Nil(t, client.ftp)

	_, err = client.ReadDir("/source-b")
	require.EqualError(t, err, "tls: first record does not look like a TLS handshake")
	assert.Equal(t, 1, dialCalls)
	assert.EqualError(t, client.connectErr, "tls: first record does not look like a TLS handshake")
}

func TestReadDirKeepsConnectionAfterNonTimeoutError(t *testing.T) {
	conn := &stubFTPConn{listErr: errors.New("permission denied")}
	client := &Client{
		ftp: conn,
	}

	_, err := client.ReadDir("/source")
	require.EqualError(t, err, "permission denied")
	assert.Equal(t, 0, conn.quitCall)
	assert.Equal(t, conn, client.ftp)
}

func TestRetrieveReconnectsAfterTimeoutClose(t *testing.T) {
	firstConn := &stubFTPConn{
		retrResp: &stubReadCloser{
			data:     bytes.NewBufferString("payload"),
			closeErr: timeoutError{},
		},
	}
	secondConn := &stubFTPConn{
		retrResp: &stubReadCloser{
			data: bytes.NewBufferString("fresh"),
		},
	}

	dialCalls := 0
	client := &Client{
		ftp: firstConn,
		dial: func(_ string, _ ...ftplib.DialOption) (ftpConn, error) {
			dialCalls++
			if dialCalls == 1 {
				return secondConn, nil
			}
			return nil, errors.New("unexpected dial")
		},
	}

	var firstDest bytes.Buffer
	err := client.Retrieve("/file.txt", &firstDest)
	require.Error(t, err)
	assert.Equal(t, "payload", firstDest.String())
	assert.Equal(t, 1, firstConn.quitCall)
	assert.Nil(t, client.ftp)

	var secondDest bytes.Buffer
	err = client.Retrieve("/file.txt", &secondDest)
	require.NoError(t, err)
	assert.Equal(t, "fresh", secondDest.String())
	assert.Equal(t, 1, dialCalls)
	assert.Equal(t, secondConn, client.ftp)
}

type listingServerOptions struct {
	mlsd        bool
	failList    bool
	failRestore bool
	home        string
}

// Exercise the real FTP library so tests cover the commands sent on the wire.
func newListingTestServer(t *testing.T, opts listingServerOptions) *ftplib.ServerConn {
	t.Helper()
	if opts.home == "" {
		opts.home = "/"
	}
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	done := make(chan struct{})
	go func() {
		defer close(done)
		control, err := listener.Accept()
		if err != nil {
			return
		}
		defer control.Close()
		_ = control.SetDeadline(time.Now().Add(10 * time.Second))
		var data net.Listener
		defer func() {
			if data != nil {
				_ = data.Close()
			}
		}()
		cwd := opts.home
		fmt.Fprint(control, "220 Test server\r\n")
		scanner := bufio.NewScanner(control)
		for scanner.Scan() {
			command, argument, _ := strings.Cut(scanner.Text(), " ")
			switch command {
			case "USER":
				fmt.Fprint(control, "230 Logged in\r\n")
			case "FEAT":
				if opts.mlsd {
					fmt.Fprint(control, "211-Features\r\n MLST type*;size*;modify*;\r\n211 End\r\n")
				} else {
					fmt.Fprint(control, "211 No features\r\n")
				}
			case "TYPE":
				fmt.Fprint(control, "200 OK\r\n")
			case "PWD":
				fmt.Fprintf(control, "257 \"%s\"\r\n", strings.ReplaceAll(cwd, `"`, `""`))
			case "CWD":
				if argument == opts.home {
					if opts.failRestore {
						fmt.Fprint(control, "550 Restore failed\r\n")
						continue
					}
					cwd = opts.home
				} else if (argument == "/complete/Game.v1 [Collection]/café" || argument == "complete/Game.v1 [Collection]/café") && cwd == opts.home {
					cwd = "/complete/Game.v1 [Collection]/café"
				} else {
					fmt.Fprint(control, "550 Invalid directory\r\n")
					continue
				}
				fmt.Fprint(control, "250 Directory changed\r\n")
			case "EPSV":
				data, err = (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
				if err != nil {
					t.Error(err)
					return
				}
				fmt.Fprintf(control, "229 Entering Extended Passive Mode (|||%d|)\r\n", data.Addr().(*net.TCPAddr).Port)
			case "MLSD", "LIST":
				if argument != "" || cwd != "/complete/Game.v1 [Collection]/café" || opts.failList {
					fmt.Fprint(control, "550 Listing failed\r\n")
					_ = data.Close()
					continue
				}
				fmt.Fprint(control, "150 Opening data connection\r\n")
				transfer, err := data.Accept()
				if err != nil {
					t.Error(err)
					return
				}
				if command == "MLSD" {
					fmt.Fprint(transfer, "type=file;size=7;modify=20261112130000; résumé.txt\r\n")
				} else {
					fmt.Fprint(transfer, "-rw-r--r-- 1 ftp ftp 7 Nov 12 2026 résumé.txt\r\n")
				}
				_ = transfer.Close()
				_ = data.Close()
				fmt.Fprint(control, "226 Complete\r\n")
			case "QUIT":
				fmt.Fprint(control, "221 Bye\r\n")
				return
			default:
				fmt.Fprint(control, "500 Unsupported command\r\n")
			}
		}
	}()
	conn, err := ftplib.Dial(listener.Addr().String(), ftplib.DialWithTimeout(5*time.Second))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = conn.Quit()
		<-done
	})
	require.NoError(t, conn.Login("test", "test"))
	return conn
}

func mustEncodeWindows1251(t *testing.T, value string) string {
	t.Helper()

	encoded, _, err := transform.String(charmap.Windows1251.NewEncoder(), value)
	require.NoError(t, err)
	return encoded
}
