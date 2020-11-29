package script

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/crazy-max/ftpgrab/v7/internal/notif/notifier"
	"github.com/hako/durafmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Client represents an active script notification object
type Client struct {
	*notifier.Notifier
	cfg  *config.NotifScript
	meta config.Meta
}

// New creates a new script notification instance
func New(config *config.NotifScript, meta config.Meta) notifier.Notifier {
	return notifier.Notifier{
		Handler: &Client{
			cfg:  config,
			meta: meta,
		},
	}
}

// Name returns notifier's name
func (c *Client) Name() string {
	return "script"
}

// Send creates and sends a slack notification with journal entries
func (c *Client) Send(jnl journal.Journal) error {
	cmd := exec.Command(c.cfg.Cmd, c.cfg.Args...)
	setSysProcAttr(cmd)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set working dir
	if c.cfg.Dir != "" {
		cmd.Dir = c.cfg.Dir
	}

	// Set env vars
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("FTPGRAB_VERSION=%s", c.meta.Version),
		fmt.Sprintf("FTPGRAB_SERVER_IP=%s", jnl.ServerHost),
		fmt.Sprintf("FTPGRAB_DEST_HOSTNAME=%s", c.meta.Hostname),
	}...)
	for idx, entry := range jnl.Entries {
		cmd.Env = append(cmd.Env, []string{
			fmt.Sprintf("FTPGRAB_JOURNAL_ENTRIES[%d]_FILE=%s", idx, entry.File),
			fmt.Sprintf("FTPGRAB_JOURNAL_ENTRIES[%d]_STATUS=%s", idx, string(entry.Status)),
			fmt.Sprintf("FTPGRAB_JOURNAL_ENTRIES[%d]_LEVEL=%s", idx, string(entry.Level)),
			fmt.Sprintf("FTPGRAB_JOURNAL_ENTRIES[%d]_TEXT=%s", idx, entry.Text),
		}...)
	}
	cmd.Env = append(cmd.Env, []string{
		fmt.Sprintf("FTPGRAB_JOURNAL_COUNT_SUCCESS=%d", jnl.Count.Success),
		fmt.Sprintf("FTPGRAB_JOURNAL_COUNT_ERROR=%d", jnl.Count.Error),
		fmt.Sprintf("FTPGRAB_JOURNAL_COUNT_SKIP=%d", jnl.Count.Skip),
		fmt.Sprintf("FTPGRAB_JOURNAL_DURATION=%s", durafmt.ParseShort(jnl.Duration).String()),
	}...)

	// Run
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, strings.TrimSpace(stderr.String()))
	}

	log.Debug().Msgf(strings.TrimSpace(stdout.String()))
	return nil
}
