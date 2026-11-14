package grabber

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// File represents a file to grab
type File struct {
	Base    string
	SrcDir  string
	DestDir string // Relative to the download output root.
	Info    os.FileInfo
}

func (c *Client) ListFiles() []File {
	var files []File

	// Iterate sources
	for _, src := range c.server.Common().Sources {
		log.Debug().Str("source", src).Msg("Listing files")

		// Check basedir
		dest := "."
		if cleanSrc := path.Clean(src); cleanSrc != "/" && *c.config.CreateBaseDir {
			var err error
			dest, err = filepath.Localize(strings.TrimPrefix(cleanSrc, "/"))
			if err != nil {
				log.Error().Err(err).Str("source", src).Msg("Cannot map source to destination")
				continue
			}
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
	if !validRemoteName(file.Name()) {
		log.Error().Str("source", srcdir).Str("name", file.Name()).Msg("Skipping unsafe remote filename")
		return nil
	}

	srcfile := path.Join(srcdir, file.Name())
	destfile := filepath.Join(destdir, file.Name())

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

// validRemoteName checks that a decoded listing entry is one local filename.
// Do not clean names before validation.
func validRemoteName(name string) bool {
	return name != "." && validLocalName(name) && !strings.ContainsAny(name, "/\\\x00")
}
