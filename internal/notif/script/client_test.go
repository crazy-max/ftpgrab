package script

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
	"github.com/crazy-max/ftpgrab/v7/internal/journal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")
	t.Setenv("FTPGRAB_EXPECT_DIR", t.TempDir())

	client := &Client{
		cfg: &config.NotifScript{
			Cmd:  os.Args[0],
			Args: []string{"-test.run=TestHelperProcess", "--", "ok"},
			Dir:  os.Getenv("FTPGRAB_EXPECT_DIR"),
		},
		meta: config.Meta{
			Version:  "v1.2.3",
			Hostname: "dest-host",
		},
	}

	err := client.Send(journal.Journal{
		ServerHost: "10.0.0.1",
		Entries: []journal.Entry{
			{
				File:   "/shows/episode.mkv",
				Status: journal.EntryStatusNeverDl,
				Level:  journal.EntryLevelSuccess,
				Text:   "downloaded",
			},
		},
		Duration: 90 * time.Second,
	})
	require.NoError(t, err)
}

func TestSendError(t *testing.T) {
	t.Setenv("GO_WANT_HELPER_PROCESS", "1")

	client := &Client{
		cfg: &config.NotifScript{
			Cmd:  os.Args[0],
			Args: []string{"-test.run=TestHelperProcess", "--", "fail"},
		},
	}

	err := client.Send(journal.Journal{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
	assert.Contains(t, err.Error(), "exit status")
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "ok":
		if os.Getenv("FTPGRAB_VERSION") != "v1.2.3" {
			fmt.Fprint(os.Stderr, "missing version")
			os.Exit(2)
		}
		if os.Getenv("FTPGRAB_SERVER_IP") != "10.0.0.1" {
			fmt.Fprint(os.Stderr, "missing server ip")
			os.Exit(2)
		}
		if os.Getenv("FTPGRAB_DEST_HOSTNAME") != "dest-host" {
			fmt.Fprint(os.Stderr, "missing hostname")
			os.Exit(2)
		}
		if os.Getenv("FTPGRAB_JOURNAL_ENTRIES[0]_FILE") != "/shows/episode.mkv" {
			fmt.Fprint(os.Stderr, "missing entry file")
			os.Exit(2)
		}
		if os.Getenv("FTPGRAB_JOURNAL_COUNT_SUCCESS") != "0" {
			fmt.Fprint(os.Stderr, "missing success count")
			os.Exit(2)
		}
		if os.Getenv("FTPGRAB_JOURNAL_DURATION") != "1 minute" {
			fmt.Fprint(os.Stderr, "missing duration")
			os.Exit(2)
		}
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(2)
		}
		if wd != os.Getenv("FTPGRAB_EXPECT_DIR") {
			fmt.Fprint(os.Stderr, "wrong dir")
			os.Exit(2)
		}
		os.Exit(0)
	case "fail":
		fmt.Fprint(os.Stderr, "boom")
		os.Exit(3)
	}
}
