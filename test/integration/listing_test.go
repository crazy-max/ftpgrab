//go:build integration

package integration

import (
	"bufio"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Normal filesystem fixtures cannot contain slash or NUL names. These servers
// send crafted listings through the real FTP and SFTP clients instead.
func TestUnsafeRemoteListings(t *testing.T) {
	for _, protocol := range []string{"ftp-list", "ftp-mlsd", "sftp"} {
		for _, tempFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/tempFirst=%t", protocol, tempFirst), func(t *testing.T) {
				fixture := &listingFixture{}
				for _, name := range []string{"safe.txt", "café-日本語.txt", "file [1].txt", ".hidden", ".ftpgrab-tmp-backup"} {
					fixture.entries = append(fixture.entries, listingInfo{name: name})
					fixture.wantFiles = append(fixture.wantFiles, "/source/"+name)
				}
				fixture.entries = append(fixture.entries, listingInfo{name: "nested", dir: true})
				fixture.wantFiles = append(fixture.wantFiles, "/source/nested/child.txt")
				unsafe := []string{".", "..", `..\outside`, `dir\file`, "nul\x00byte"}
				if protocol != "sftp" {
					unsafe = append(unsafe, "../outside", "/absolute", "dir/file", "nested/../escape")
				} else {
					// pkg/sftp applies path.Base before exposing FileInfo.Name.
					fixture.entries = append(fixture.entries, listingInfo{name: "../normalized.txt"})
					fixture.wantFiles = append(fixture.wantFiles, "/source/normalized.txt")
				}
				unsafe = append(unsafe, unsafeHostNames()...)
				for _, name := range unsafe {
					fixture.entries = append(fixture.entries, listingInfo{name: name}, listingInfo{name: name, dir: true})
				}

				var srv *config.Server
				if protocol == "sftp" {
					srv = serveListingSFTP(t, fixture)
				} else {
					srv = serveListingFTP(t, fixture, protocol == "ftp-mlsd")
				}
				base := t.TempDir()
				sentinel := filepath.Join(base, "outside")
				require.NoError(t, os.WriteFile(sentinel, []byte("untouched"), 0o600))
				dl := (&config.Download{}).GetDefaults()
				dl.Output = filepath.Join(base, "output")
				dl.TempFirst = new(tempFirst)
				require.NoError(t, os.Mkdir(dl.Output, 0o755))
				client := openGrabber(t, dl, nil, srv)
				files := client.ListFiles()
				var sources []string
				for _, file := range files {
					sources = append(sources, path.Join(file.SrcDir, file.Info.Name()))
				}
				require.ElementsMatch(t, fixture.wantFiles, sources)
				result := client.Grab(files)
				assert.Equal(t, len(fixture.wantFiles), result.Count.Success)
				assert.Zero(t, result.Count.Error)
				for _, source := range fixture.wantFiles {
					data, err := os.ReadFile(filepath.Join(dl.Output, strings.TrimPrefix(source, "/source/")))
					require.NoError(t, err)
					assert.Equal(t, "payload", string(data))
				}
				data, err := os.ReadFile(sentinel)
				require.NoError(t, err)
				assert.Equal(t, "untouched", string(data))
				fixture.mu.Lock()
				defer fixture.mu.Unlock()
				assert.ElementsMatch(t, []string{"/source", "/source/nested"}, fixture.lists)
				assert.ElementsMatch(t, fixture.wantFiles, fixture.reads)
			})
		}
	}
}

func TestDestinationLocalStaging(t *testing.T) {
	for _, protocol := range []string{"ftp", "sftp"} {
		t.Run(protocol, func(t *testing.T) {
			output := t.TempDir()
			require.NoError(t, os.Mkdir(filepath.Join(output, "nested"), 0o755))
			if os.Geteuid() >= 0 {
				require.NoError(t, os.Chmod(output, 0o555))
				t.Cleanup(func() { _ = os.Chmod(output, 0o755) })
			}
			fixture := &listingFixture{
				entries:   listingInfos{listingInfo{name: "nested", dir: true}},
				wantFiles: []string{"/source/nested/child.txt"},
			}
			var srv *config.Server
			if protocol == "sftp" {
				srv = serveListingSFTP(t, fixture)
			} else {
				srv = serveListingFTP(t, fixture, false)
			}
			dl := (&config.Download{}).GetDefaults()
			dl.Output = output
			dl.TempFirst = new(false)
			client := openGrabber(t, dl, nil, srv)
			// Catch eager staging even when root privileges bypass the mode above.
			entries, err := os.ReadDir(output)
			require.NoError(t, err)
			require.Len(t, entries, 1)
			assert.Equal(t, "nested", entries[0].Name())

			result := client.Grab(client.ListFiles())
			assert.Equal(t, 1, result.Count.Success)
			assert.Zero(t, result.Count.Error)
			client.Close()
			data, err := os.ReadFile(filepath.Join(output, "nested", "child.txt"))
			require.NoError(t, err)
			assert.Equal(t, "payload", string(data))
			assertNoTemporaryDirectories(t, output)
			if os.Geteuid() >= 0 {
				info, err := os.Stat(output)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0o555), info.Mode().Perm())
			}
		})
	}
}

type listingInfo struct {
	name string
	dir  bool
}

func (f listingInfo) Name() string       { return f.name }
func (f listingInfo) Size() int64        { return 7 }
func (f listingInfo) ModTime() time.Time { return time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC) }
func (f listingInfo) IsDir() bool        { return f.dir }
func (f listingInfo) Sys() any           { return nil }
func (f listingInfo) Mode() os.FileMode {
	if f.dir {
		return os.ModeDir | 0o755
	}
	return 0o644
}

type listingInfos []os.FileInfo

func (entries listingInfos) ListAt(dest []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(entries)) {
		return 0, io.EOF
	}
	n := copy(dest, entries[offset:])
	if n < len(dest) {
		return n, io.EOF
	}
	return n, nil
}

type listingFixture struct {
	mu        sync.Mutex
	entries   listingInfos
	wantFiles []string
	lists     []string
	reads     []string
}

func (f *listingFixture) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch r.Method {
	case "List":
		f.lists = append(f.lists, r.Filepath)
		switch r.Filepath {
		case "/source":
			return f.entries, nil
		case "/source/nested":
			return listingInfos{listingInfo{name: "child.txt"}}, nil
		}
	case "Stat":
		if slices.Contains(f.wantFiles, r.Filepath) {
			return listingInfos{listingInfo{name: path.Base(r.Filepath)}}, nil
		}
	}
	return nil, os.ErrNotExist
}

func (f *listingFixture) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reads = append(f.reads, r.Filepath)
	if slices.Contains(f.wantFiles, r.Filepath) {
		return strings.NewReader("payload"), nil
	}
	return nil, os.ErrNotExist
}

func serveListingSFTP(t *testing.T, fixture *listingFixture) *config.Server {
	t.Helper()
	_, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	signer, err := ssh.NewSignerFromKey(private)
	require.NoError(t, err)
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	done := make(chan struct{})
	t.Cleanup(func() { _ = listener.Close(); <-done })
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
		cfg := &ssh.ServerConfig{NoClientAuth: true}
		cfg.AddHostKey(signer)
		sshConn, channels, requests, err := ssh.NewServerConn(conn, cfg)
		if !assert.NoError(t, err) {
			return
		}
		defer sshConn.Close()
		go ssh.DiscardRequests(requests)
		for request := range channels {
			channel, requests, err := request.Accept()
			if !assert.NoError(t, err) {
				return
			}
			for req := range requests {
				if req.Type == "subsystem" {
					_ = req.Reply(true, nil)
					server := sftp.NewRequestServer(channel, sftp.Handlers{FileGet: fixture, FileList: fixture})
					err := server.Serve()
					_ = server.Close()
					if err != io.EOF {
						assert.NoError(t, err)
					}
					return
				}
				_ = req.Reply(false, nil)
			}
		}
	}()
	cfg := (&config.ServerSFTP{}).GetDefaults()
	cfg.Host, cfg.Port = "127.0.0.1", listener.Addr().(*net.TCPAddr).Port
	cfg.Sources = []string{"/source"}
	cfg.KnownHostsFile = filepath.Join(t.TempDir(), "known_hosts")
	line := knownhosts.Line([]string{knownhosts.Normalize(listener.Addr().String())}, signer.PublicKey())
	require.NoError(t, os.WriteFile(cfg.KnownHostsFile, []byte(line+"\n"), 0o600))
	return &config.Server{SFTP: cfg}
}

func serveListingFTP(t *testing.T, fixture *listingFixture, mlsd bool) *config.Server {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	done := make(chan struct{})
	t.Cleanup(func() { _ = listener.Close(); <-done })
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
		var data *net.TCPListener
		defer func() {
			if data != nil {
				_ = data.Close()
			}
		}()
		cwd := "/"
		fmt.Fprint(conn, "220 Listing fixture\r\n")
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			command, argument, _ := strings.Cut(scanner.Text(), " ")
			switch command {
			case "USER":
				fmt.Fprint(conn, "230 Logged in\r\n")
			case "FEAT":
				if mlsd {
					fmt.Fprint(conn, "211-Features\r\n MLST type*;size*;modify*;\r\n211 End\r\n")
				} else {
					fmt.Fprint(conn, "211 No features\r\n")
				}
			case "TYPE":
				fmt.Fprint(conn, "200 OK\r\n")
			case "PWD":
				fmt.Fprintf(conn, "257 %q\r\n", cwd)
			case "CWD":
				cwd = argument
				fmt.Fprint(conn, "250 Directory changed\r\n")
			case "EPSV":
				dataListener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
				if !assert.NoError(t, err) {
					return
				}
				data = dataListener.(*net.TCPListener)
				_ = data.SetDeadline(time.Now().Add(10 * time.Second))
				fmt.Fprintf(conn, "229 Entering Extended Passive Mode (|||%d|)\r\n", data.Addr().(*net.TCPAddr).Port)
			case "LIST", "MLSD", "RETR":
				fmt.Fprint(conn, "150 Opening data connection\r\n")
				transfer, err := data.Accept()
				if !assert.NoError(t, err) {
					return
				}
				_ = transfer.SetDeadline(time.Now().Add(10 * time.Second))
				if command == "RETR" {
					_, err = fixture.Fileread(&sftp.Request{Filepath: argument})
					if assert.NoError(t, err) {
						fmt.Fprint(transfer, "payload")
					}
				} else {
					entries, err := fixture.Filelist(&sftp.Request{Method: "List", Filepath: cwd})
					if assert.NoError(t, err) {
						for _, entry := range entries.(listingInfos) {
							kind, mode := "file", "-rw-r--r--"
							if entry.IsDir() {
								kind, mode = "dir", "drwxr-xr-x"
							}
							if command == "MLSD" {
								fmt.Fprintf(transfer, "type=%s;size=7;modify=20240102000000; %s\r\n", kind, entry.Name())
							} else {
								fmt.Fprintf(transfer, "%s 1 ftp ftp 7 Jan 2 2024 %s\r\n", mode, entry.Name())
							}
						}
					}
				}
				_ = transfer.Close()
				_ = data.Close()
				fmt.Fprint(conn, "226 Complete\r\n")
			case "QUIT":
				fmt.Fprint(conn, "221 Bye\r\n")
				return
			default:
				fmt.Fprint(conn, "500 Unsupported command\r\n")
			}
		}
	}()
	cfg := (&config.ServerFTP{}).GetDefaults()
	cfg.Host, cfg.Port = "127.0.0.1", listener.Addr().(*net.TCPAddr).Port
	cfg.Sources = []string{"/source"}
	return &config.Server{FTP: cfg}
}
