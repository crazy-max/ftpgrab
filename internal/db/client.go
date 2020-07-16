package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ftpgrab/ftpgrab/v7/internal/model"
	"github.com/ftpgrab/ftpgrab/v7/pkg/utl"
	"github.com/rs/zerolog/log"
	bolt "go.etcd.io/bbolt"
)

// Client represents an active db object
type Client struct {
	*bolt.DB
	cfg    *model.Db
	bucket string
}

type entry struct {
	File string    `json:"file"`
	Date time.Time `json:"date"`
	Size int64     `json:"size"`
}

// New creates new db instance
func New(cfg *model.Db) (c *Client, err error) {
	var db *bolt.DB
	var bucket = "ftpgrab"

	if cfg == nil || len(cfg.Path) == 0 {
		return &Client{
			cfg:    cfg,
			bucket: bucket,
		}, nil
	}

	db, err = bolt.Open(cfg.Path, 0600, &bolt.Options{
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
	return c.cfg != nil && len(c.cfg.Path) > 0
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
	if !c.Enabled() {
		return false
	}

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
	if !c.Enabled() {
		return nil
	}

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
