package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ftpgrab/ftpgrab/internal/config"
	"github.com/ftpgrab/ftpgrab/internal/utl"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// Client represents an active db object
type Client struct {
	*bolt.DB
	fg     *config.Configuration
	bucket string
}

type entry struct {
	File string    `json:"file"`
	Date time.Time `json:"date"`
	Size int64     `json:"size"`
}

// New creates new db instance
func New(cfg *config.Configuration) (c *Client, err error) {
	var db *bolt.DB
	var bucket = "ftpgrab"

	if !cfg.Download.HashEnabled {
		return c, nil
	}

	db, err = bolt.Open(fmt.Sprintf("%s.db", utl.Basename(cfg.File.Name())), 0600, &bolt.Options{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	if err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	}); err != nil {
		return nil, err
	}

	if err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		stats := b.Stats()
		log.Debug().Msgf("%d entries found in database", stats.KeyN)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("cannot count entries in database, %v", err)
	}

	return &Client{db, cfg, bucket}, nil
}

// Enabled verifies if db is enabled
func (c *Client) Enabled() bool {
	return c.fg.Download.HashEnabled
}

// Close closes db connection
func (c *Client) Close() error {
	if !c.Enabled() {
		return nil
	}
	return c.DB.Close()
}

// HasHash checks if hash is present for a file in db
func (c *Client) HasHash(base string, source string, file os.FileInfo) bool {
	exists := false
	filename := strings.TrimPrefix(path.Join(source, file.Name()), base)
	hash := utl.Hash(filename)

	_ = c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucket))
		if entryBytes := b.Get([]byte(hash)); entryBytes != nil {
			exists = true
		}
		return nil
	})

	return exists
}

// PutHash add hash in db for a given file
func (c *Client) PutHash(base string, source string, file os.FileInfo) error {
	filename := strings.TrimPrefix(path.Join(source, file.Name()), base)
	hash := utl.Hash(filename)

	entryBytes, _ := json.Marshal(entry{
		File: filename,
		Size: file.Size(),
		Date: time.Now(),
	})

	err := c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c.bucket))
		return b.Put([]byte(hash), entryBytes)
	})

	return err
}
