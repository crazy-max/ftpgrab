package grabber

import (
	"os"
	"path"

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

		files = append(files, c.readDir(src, src, dest)...)
	}

	return files
}

func (c *Client) readDir(base string, srcdir string, destdir string) []File {
	var files []File

	items, err := c.server.ReadDir(srcdir)
	if err != nil {
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
		return c.readDir(base, srcfile, destfile)
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
