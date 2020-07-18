package notifier

import "github.com/crazy-max/ftpgrab/v7/internal/journal"

// Handler is a notifier interface
type Handler interface {
	Name() string
	Send(jnl journal.Client) error
}

// Notifier represents an active notifier object
type Notifier struct {
	Handler
}
