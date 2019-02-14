package model

// Db holds data necessary for Database configuration
type Db struct {
	Enable bool   `yaml:"enable,omitempty"`
	Path   string `yaml:"path,omitempty"`
}
