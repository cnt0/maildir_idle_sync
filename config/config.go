package config

import (
	"bytes"
	"os/exec"

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
	PasswordCommand PasswordCommand
	Mailboxes       []Mailbox
	UpdateEvents    *events.UpdateEvents
}

// PasswordCommand ...
type PasswordCommand string

// UnmarshalText ...
func (cmd *PasswordCommand) UnmarshalText(text []byte) error {
	args := []string{}
	for _, tok := range bytes.Split(text, []byte{' '}) {
		args = append(args, string(tok))
	}

	// it's ok if PasswordCommand is not given
	if len(args) == 0 {
		return nil
	}

	output, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		return err
	}
	*cmd = PasswordCommand(bytes.TrimSpace(output))
	return nil
}
