//go:build !windows

package grabber

import "path/filepath"

func validLocalName(name string) bool {
	return filepath.IsLocal(name)
}
