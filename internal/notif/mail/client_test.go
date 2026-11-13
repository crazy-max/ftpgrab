package mail

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/http"
	"net/http/httptest"
	stdmail "net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendSMTPRegression(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)
	tests := []struct {
		name      string
		options   smtpServerOptions
		wantError string
	}{
		{
			name: "plain relay",
		},
		{
			name: "advertised auth without credentials",
			options: smtpServerOptions{
				advertiseAuth: true,
			},
		},
		{
			name: "relay rejects NOOP",
			options: smtpServerOptions{
				closeOnNoop: true,
			},
		},
		{
			name: "relay closes after accepting DATA",
			options: smtpServerOptions{
				closeAfterData: true,
			},
		},
		{
			name: "delivery rejected",
			options: smtpServerOptions{
				rejectData: true,
			},
			wantError: "cannot send mail notification",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, done := startSMTPServer(t, tt.options)
			client := newTestClient(t, addr)
			err := client.Send(journal.Journal{})
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
			} else {
				require.NoError(t, err)
			}
			result := waitSMTPServer(t, done)
			require.NoError(t, result.err)
			assert.Equal(t, hostname, result.helo)
			assert.False(t, result.authSeen)
			assert.False(t, result.noopSeen)
			assert.NotEmpty(t, result.data)
		})
	}
}

func TestSendMailOptions(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)
	tests := []struct {
		name       string
		localName  string
		recipients []string
		wantHelo   string
	}{
		{
			name:       "default HELO and single recipient",
			recipients: []string{"ops@example.com"},
			wantHelo:   hostname,
		},
		{
			name:       "custom HELO and multiple recipients",
			localName:  "ftpgrab.example.com",
			recipients: []string{"ops@example.com", "dev@example.com"},
			wantHelo:   "ftpgrab.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, done := startSMTPServer(t, smtpServerOptions{})
			client := newTestClient(t, addr)
			client.cfg.LocalName = tt.localName
			client.cfg.To = tt.recipients
			require.NoError(t, client.Send(journal.Journal{}))
			result := waitSMTPServer(t, done)
			require.NoError(t, result.err)
			assert.Equal(t, tt.wantHelo, result.helo)
			message, err := stdmail.ReadMessage(strings.NewReader(result.data))
			require.NoError(t, err)
			to, err := message.Header.AddressList("To")
			require.NoError(t, err)
			require.Len(t, to, len(tt.recipients))
			require.Len(t, result.to, len(tt.recipients))
			for i, recipient := range tt.recipients {
				assert.Equal(t, recipient, to[i].Address)
				assert.Equal(t, "RCPT TO:<"+recipient+">", result.to[i])
			}
		})
	}
}

func TestSendRejectsRecipient(t *testing.T) {
	addr, done := startSMTPServer(t, smtpServerOptions{
		rejectRecipient: "dev@example.com",
	})
	client := newTestClient(t, addr)
	client.cfg.To = []string{"ops@example.com", "dev@example.com"}
	err := client.Send(journal.Journal{})
	require.ErrorContains(t, err, "cannot send mail notification")
	assert.Contains(t, err.Error(), "dev@example.com")
	result := waitSMTPServer(t, done)
	require.NoError(t, result.err)
	assert.Len(t, result.to, 2)
	assert.Empty(t, result.data)
}

func TestSendReport(t *testing.T) {
	addr, done := startSMTPServer(t, smtpServerOptions{})
	client := newTestClient(t, addr)
	jnl := journal.Journal{
		ServerHost: "ftp.example.com",
		Duration:   90 * time.Second,
		Entries: []journal.Entry{
			{
				File:  "/shows/episode.mkv",
				Level: journal.EntryLevelSuccess,
				Text:  "downloaded successfully",
			},
		},
	}
	jnl.Count.Success = 2
	jnl.Count.Skip = 1
	jnl.Count.Error = 3
	require.NoError(t, client.Send(jnl))
	result := waitSMTPServer(t, done)
	require.NoError(t, result.err)
	assert.Equal(t, "MAIL FROM:<ftpgrab@example.com>", result.from)
	assert.Equal(t, []string{"RCPT TO:<ops@example.com>"}, result.to)
	message, err := stdmail.ReadMessage(strings.NewReader(result.data))
	require.NoError(t, err)
	from, err := message.Header.AddressList("From")
	require.NoError(t, err)
	require.Len(t, from, 1)
	assert.Equal(t, "FTPGrab", from[0].Name)
	assert.Equal(t, client.cfg.From, from[0].Address)
	to, err := message.Header.AddressList("To")
	require.NoError(t, err)
	require.Len(t, to, 1)
	assert.Equal(t, client.cfg.To[0], to[0].Address)
	assert.Equal(t, "FTPGrab report for ftp.example.com on destination", message.Header.Get("Subject"))
	assert.NotEmpty(t, message.Header.Get("Message-ID"))
	assert.NotEmpty(t, message.Header.Get("Date"))
	mediaType, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	require.NoError(t, err)
	assert.Equal(t, "multipart/alternative", mediaType)
	reader := multipart.NewReader(message.Body, params["boundary"])
	for _, wantType := range []string{"text/plain", "text/html"} {
		part, err := reader.NextPart()
		require.NoError(t, err)
		partType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		require.NoError(t, err)
		assert.Equal(t, wantType, partType)
		var bodyReader io.Reader = part
		if part.Header.Get("Content-Transfer-Encoding") == "base64" {
			bodyReader = base64.NewDecoder(base64.StdEncoding, part)
		} else if part.Header.Get("Content-Transfer-Encoding") == "quoted-printable" {
			bodyReader = quotedprintable.NewReader(part)
		}
		body, err := io.ReadAll(bodyReader)
		require.NoError(t, err)
		assert.Contains(t, string(body), "report")
		if wantType == "text/html" {
			assert.Contains(t, string(body), "/shows/episode.mkv")
			assert.Contains(t, string(body), "downloaded successfully")
		} else {
			assert.Contains(t, string(body), "*2* files have been download successfully, *1* have been skipped and *3* errors occurred in 1 minute.")
		}
		assert.Contains(t, string(body), "success")
		assert.Contains(t, string(body), "1 minute")
		assert.Contains(t, string(body), "Thanks for your support")
	}
	_, err = reader.NextPart()
	require.ErrorIs(t, err, io.EOF)
}

func TestSendTLSAndAuth(t *testing.T) {
	certServer := httptest.NewTLSServer(http.NotFoundHandler())
	serverTLS := certServer.TLS.Clone()
	certServer.Close()
	tests := []struct {
		name       string
		ssl        bool
		skipVerify bool
		wantError  bool
	}{
		{
			name:       "STARTTLS with credential files",
			skipVerify: true,
		},
		{
			name:       "implicit TLS with credential files",
			ssl:        true,
			skipVerify: true,
		},
		{
			name:      "STARTTLS verifies certificate",
			wantError: true,
		},
		{
			name:      "implicit TLS verifies certificate",
			ssl:       true,
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, done := startSMTPServer(t, smtpServerOptions{
				advertiseAuth: true,
				implicitTLS:   tt.ssl,
				tlsConfig:     serverTLS,
			})
			client := newTestClient(t, addr)
			client.cfg.SSL = new(tt.ssl)
			client.cfg.InsecureSkipVerify = new(tt.skipVerify)
			client.cfg.UsernameFile = filepath.Join(t.TempDir(), "username")
			client.cfg.PasswordFile = filepath.Join(t.TempDir(), "password")
			require.NoError(t, os.WriteFile(client.cfg.UsernameFile, []byte("smtp-user"), 0o600))
			require.NoError(t, os.WriteFile(client.cfg.PasswordFile, []byte("smtp-password"), 0o600))
			err := client.Send(journal.Journal{})
			if tt.wantError {
				require.ErrorContains(t, err, "cannot send mail notification")
			} else {
				require.NoError(t, err)
			}
			result := waitSMTPServer(t, done)
			if tt.wantError {
				assert.Empty(t, result.data)
				return
			}
			require.NoError(t, result.err)
			assert.True(t, result.authSeen)
			assert.Equal(t, "smtp-user", result.username)
			assert.Equal(t, "smtp-password", result.password)
			assert.NotEmpty(t, result.data)
		})
	}
}

type smtpServerOptions struct {
	advertiseAuth   bool
	closeAfterData  bool
	closeOnNoop     bool
	rejectData      bool
	rejectRecipient string
	implicitTLS     bool
	tlsConfig       *tls.Config
}

type smtpServerResult struct {
	err      error
	helo     string
	authSeen bool
	noopSeen bool
	username string
	password string
	from     string
	to       []string
	data     string
}

func startSMTPServer(t *testing.T, options smtpServerOptions) (string, <-chan smtpServerResult) {
	t.Helper()
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })
	done := make(chan smtpServerResult, 1)
	go func() {
		var result smtpServerResult
		defer func() { done <- result }()
		conn, err := listener.Accept()
		if err != nil {
			result.err = err
			return
		}
		defer conn.Close()
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			result.err = err
			return
		}
		if options.implicitTLS {
			conn = tls.Server(conn, options.tlsConfig)
		}
		tp := textproto.NewConn(conn)
		defer func() { _ = tp.Close() }()
		write := func(line string) { _ = tp.PrintfLine("%s", line) }
		write("220 localhost ESMTP")
		for {
			line, err := tp.ReadLine()
			if err != nil {
				if err != io.EOF {
					result.err = err
				}
				return
			}
			upperLine := strings.ToUpper(line)
			switch {
			case strings.HasPrefix(upperLine, "EHLO "), strings.HasPrefix(upperLine, "HELO "):
				_, result.helo, _ = strings.Cut(line, " ")
				if options.tlsConfig != nil && !options.implicitTLS {
					write("250-localhost")
					write("250-STARTTLS")
				} else {
					write("250-localhost")
				}
				if options.advertiseAuth {
					write("250 AUTH PLAIN")
				} else {
					write("250 OK")
				}
			case upperLine == "STARTTLS":
				write("220 Ready for TLS")
				conn = tls.Server(conn, options.tlsConfig)
				tp = textproto.NewConn(conn)
			case upperLine == "NOOP":
				result.noopSeen = true
				if options.closeOnNoop {
					return
				}
				write("250 OK")
			case strings.HasPrefix(upperLine, "AUTH PLAIN "):
				result.authSeen = true
				encoded := strings.TrimPrefix(line, "AUTH PLAIN ")
				decoded, err := base64.StdEncoding.DecodeString(encoded)
				if err != nil {
					result.err = err
					return
				}
				fields := strings.Split(string(decoded), "\x00")
				if len(fields) == 3 {
					result.username, result.password = fields[1], fields[2]
				}
				write("235 authenticated")
			case strings.HasPrefix(upperLine, "MAIL FROM:"):
				result.from = line
				write("250 OK")
			case strings.HasPrefix(upperLine, "RCPT TO:"):
				result.to = append(result.to, line)
				if options.rejectRecipient != "" && line == "RCPT TO:<"+options.rejectRecipient+">" {
					write("550 recipient rejected")
					continue
				}
				write("250 OK")
			case upperLine == "DATA":
				write("354 end data with <CR><LF>.<CR><LF>")
				data, err := io.ReadAll(tp.DotReader())
				if err != nil {
					result.err = err
					return
				}
				result.data = string(data)
				if options.rejectData {
					write("550 delivery rejected")
				} else {
					write("250 OK")
				}
				if options.closeAfterData {
					return
				}
			case upperLine == "RSET":
				write("250 OK")
			case upperLine == "QUIT":
				write("221 bye")
				return
			default:
				write("500 unsupported command")
			}
		}
	}()
	return listener.Addr().String(), done
}

func waitSMTPServer(t *testing.T, done <-chan smtpServerResult) smtpServerResult {
	t.Helper()
	select {
	case result := <-done:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for SMTP server")
		return smtpServerResult{}
	}
}

func newTestClient(t *testing.T, addr string) *Client {
	t.Helper()
	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	portNum, err := strconv.Atoi(port)
	require.NoError(t, err)
	cfg := (&config.NotifMail{}).GetDefaults()
	cfg.Host = host
	cfg.Port = portNum
	cfg.From = "ftpgrab@example.com"
	cfg.To = []string{"ops@example.com"}
	return &Client{
		cfg: cfg,
		meta: config.Meta{
			Name:     "FTPGrab",
			Hostname: "destination",
		},
	}
}
