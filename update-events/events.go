package events

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"

	"github.com/emersion/go-imap/client"
)

// Command ...
type Command struct {
	exec.Cmd
}

// UnmarshalText ...
func (c *Command) UnmarshalText(text []byte) error {
	args := []string{}
	for _, tok := range bytes.Split(text, []byte{' '}) {
		args = append(args, string(tok))
	}
	if len(args) == 0 {
		return fmt.Errorf("Incorrect command")
	}
	c.Cmd = *exec.Command(args[0], args[1:]...)
	return nil
}

// ErrNoEventHandler ...
var ErrNoEventHandler = errors.New("Event handler is not given")

// UpdateEvents ...
type UpdateEvents struct {
	Default   *Command
	OnStatus  *Command
	OnMailbox *Command
	OnMessage *Command
	OnExpunge *Command
}

// Handle ...
func (ev *UpdateEvents) Handle(update client.Update) error {
	switch interface{}(update).(type) {
	case *client.StatusUpdate:
		log.Println("Got status update")
		if ev.OnStatus != nil {
			return ev.OnStatus.Run()
		}
		// type switch doesn't support fallthrough FeelsBadMan
		if ev.Default != nil {
			return ev.Default.Run()
		}
		return ErrNoEventHandler
	case *client.MailboxUpdate:
		log.Println("Got mailbox update")
		if ev.OnMailbox != nil {
			log.Println("RUN", ev.OnMailbox)
			return ev.OnMailbox.Run()
		}
		if ev.Default != nil {
			log.Println("RUN", ev.Default)
			return ev.Default.Run()
		}
		return ErrNoEventHandler
	case *client.MessageUpdate:
		log.Println("Got message update")
		if ev.OnMessage != nil {
			return ev.OnMessage.Run()
		}
		if ev.Default != nil {
			return ev.Default.Run()
		}
		return ErrNoEventHandler
	case *client.ExpungeUpdate:
		log.Println("Got expunge update")
		if ev.OnExpunge != nil {
			return ev.OnExpunge.Run()
		}
		if ev.Default != nil {
			return ev.Default.Run()
		}
		return ErrNoEventHandler
	default:
		return fmt.Errorf("weird update tbh")

	}

}
