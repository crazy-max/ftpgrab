package grabber

import (
	"os"
	"path"

	"github.com/crazy-max/ftpgrab/v7/internal/server/ftp"
	"github.com/rs/zerolog/log"
)

// File represents a file to grab
type File struct {
	Base    string
	SrcDir  string
	DestDir string
	Info    os.FileInfo
}

func (c *Client) ListFiles() []File {
	var files []File

	// Iterate sources
	for _, src := range c.server.Common().Sources {
		log.Debug().Str("source", src).Msg("Listing files")

		// Check basedir
		dest := c.config.Output
		if src != "/" && *c.config.CreateBaseDir {
			dest = path.Join(dest, src)
		}

		files = append(files, c.readDir(src, src, dest, 0)...)
	}

	return files
}

func (c *Client) readDir(base string, srcdir string, destdir string, retry int) []File {
	var files []File

	items, err := c.server.ReadDir(srcdir)
	if err != nil {
		// If list operation timed out, reconnect and retry
		if err == ftp.ErrListTimeout {
			retry++
			log.Warn().Str("source", base).Msgf("List operation timed out, retry %d/%d", retry, c.config.Retry)
			if retry == c.config.Retry {
				log.Error().Str("source", base).Msgf("Cannot read directory %s after %d retries", srcdir, retry)
				return []File{}
			}
			log.Warn().Str("source", base).Msg("Reconnecting to server after list timeout")
			if reconnectErr := c.reconnect(); reconnectErr != nil {
				log.Error().Err(reconnectErr).Str("source", base).Msg("Cannot reconnect to server")
				return []File{}
			}
			// Retry with incremented retry count
			return c.readDir(base, srcdir, destdir, retry)
		}
		log.Error().Err(err).Str("source", base).Msgf("Cannot read directory %s", srcdir)
		return []File{}
	}

	for _, item := range items {
		files = append(files, c.readFile(base, srcdir, destdir, item)...)
	}

	return files
}

func (c *Client) readFile(base string, srcdir string, destdir string, file os.FileInfo) []File {
	srcfile := path.Join(srcdir, file.Name())
	destfile := path.Join(destdir, file.Name())

	if file.IsDir() {
		return c.readDir(base, srcfile, destfile, 0)
	}

	return []File{
		{
			Base:    base,
			SrcDir:  srcdir,
			DestDir: destdir,
			Info:    file,
		},
	}
}
