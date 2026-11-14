package grabber

import (
	"path/filepath"
	"strings"
)

func validLocalName(name string) bool {
	// Win32 strips trailing dots and spaces, which can alias other entries.
	return filepath.IsLocal(name) && strings.TrimRight(name, " .") == name
}
