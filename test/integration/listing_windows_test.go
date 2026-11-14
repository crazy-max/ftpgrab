//go:build integration

package integration

func unsafeHostNames() []string {
	return []string{".. ", ". ", "...", "safe.txt.", "safe.txt ", "NUL", "CON", "C:file", "file:stream"}
}
