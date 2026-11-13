package ftp

import (
	stderrors "errors"
	"io"
	"net"
	"os"
	"slices"
	"sync"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/rs/zerolog/log"
)

// pool lends an exclusive FTP session to each operation. The grabber bounds the
// number of concurrent operations, so connections are created only as needed.
type pool struct {
	common      config.ServerCommon
	connectFunc func() (*Client, error)
	mu          sync.Mutex
	available   *sync.Cond
	sessions    []*Client
	idle        []*Client
	connecting  bool
	limited     bool
	closed      bool
}

func newPool(client *Client) *pool {
	p := &pool{
		common:      client.Common(),
		connectFunc: client.newSession,
		sessions:    []*Client{client},
		idle:        []*Client{client},
	}
	p.available = sync.NewCond(&p.mu)
	return p
}

func (p *pool) Common() config.ServerCommon {
	return p.common
}

func (p *pool) acquire() (*Client, error) {
	p.mu.Lock()
	for {
		if p.closed {
			p.mu.Unlock()
			return nil, net.ErrClosed
		}
		if len(p.idle) > 0 {
			session := p.idle[len(p.idle)-1]
			p.idle = p.idle[:len(p.idle)-1]
			p.mu.Unlock()
			return session, nil
		}
		if p.limited || p.connecting {
			p.available.Wait()
			continue
		}
		// Dial outside the lock so active sessions can finish and be reused.
		// Permit only one pending connection attempt to avoid a login burst.
		p.connecting = true
		p.mu.Unlock()
		session, err := p.connectFunc()
		p.mu.Lock()
		p.connecting = false
		p.available.Broadcast()
		if p.closed {
			p.mu.Unlock()
			if err == nil {
				p.release(session)
			}
			return nil, net.ErrClosed
		}
		if err != nil {
			p.limited = true
			log.Warn().Err(err).Int("sessions", len(p.sessions)).Msg("Cannot open additional FTP session, continuing with fewer connections")
		} else {
			p.sessions = append(p.sessions, session)
			p.idle = append(p.idle, session)
		}
	}
}

func (p *pool) release(session *Client) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		if err := session.Close(); err != nil {
			log.Warn().Err(err).Msg("Cannot close returned FTP session")
		}
		return
	}
	if session.connectErr != nil && len(p.sessions) > 1 {
		// A failed reconnect is cached by Client. Retire that session rather
		// than repeatedly lending it while healthy sessions remain. Keep the
		// last session's cached error to fail fast for the rest of the run.
		p.sessions = slices.DeleteFunc(p.sessions, func(c *Client) bool {
			return c == session
		})
		p.limited = false
	} else {
		p.idle = append(p.idle, session)
	}
	p.available.Broadcast()
	p.mu.Unlock()
}

func (p *pool) ReadDir(dir string) ([]os.FileInfo, error) {
	session, err := p.acquire()
	if err != nil {
		return nil, err
	}
	defer p.release(session)
	return session.ReadDir(dir)
}

func (p *pool) Retrieve(filename string, dest io.Writer) error {
	session, err := p.acquire()
	if err != nil {
		return err
	}
	defer p.release(session)
	return session.Retrieve(filename, dest)
}

func (p *pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.available.Broadcast()
	idle := p.idle
	p.sessions = nil
	p.idle = nil
	p.mu.Unlock()
	// Active operations and pending connections close their own sessions when
	// they return. Signal handling must not wait for a long-running transfer.
	var errs []error
	for _, session := range idle {
		errs = append(errs, session.Close())
	}
	return stderrors.Join(errs...)
}
