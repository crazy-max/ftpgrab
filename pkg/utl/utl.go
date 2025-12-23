package utl

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// GetEnv retrieves the value of the environment variable named by the key
// or fallback if not found
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// GetSecret retrieves secret's value from plaintext or filename if defined
func GetSecret(plaintext, filename string) (string, error) {
	if plaintext != "" {
		return plaintext, nil
	} else if filename != "" {
		b, err := os.ReadFile(filename)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", nil
}

// Exists reports whether the named file or directory exists
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// Hash a string using SHA256
func Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Basename returns trailing name component of path
func Basename(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[0 : len(base)-len(ext)]
}

// MatchString reports whether a string s
// contains any match of a regular expression.
func MatchString(exp string, s string) bool {
	re, err := regexp.Compile(exp)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// NewFalse returns a false bool pointer
func NewFalse() *bool {
	b := false
	return &b
}

// NewTrue returns a true bool pointer
func NewTrue() *bool {
	b := true
	return &b
}

// NewDuration returns a duration pointer
func NewDuration(duration time.Duration) *time.Duration {
	return &duration
}
