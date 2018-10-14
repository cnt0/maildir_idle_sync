package config

import (
	"github.com/cnt0/maildir_idle_sync/imap"
	events "github.com/cnt0/maildir_idle_sync/update-events"
)

// Config ...
type Config struct {
	Account      []Account
	UpdateEvents *events.UpdateEvents
}

// Mailbox ...
type Mailbox struct {
	Name         string
	UpdateEvents *events.UpdateEvents
}

// Account ...
type Account struct {
	*imap.Account
	Mailboxes    []Mailbox
	UpdateEvents *events.UpdateEvents
}
