package server

import (
	"io"
	"os"

	"github.com/crazy-max/ftpgrab/v7/internal/config"
)

// Handler is a server interface
type Handler interface {
	Common() config.ServerCommon
	ReadDir(source string) ([]os.FileInfo, error)
	Retrieve(path string, dest io.Writer) error
	Close() error
}

// Client represents an active server object
type Client struct {
	Handler
}
