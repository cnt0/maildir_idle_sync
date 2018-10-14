package imap

import (
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"
)

// Account ...
type Account struct {
	Name string
	Host string
	User string
	Pass string
}

// Login ...
func (acc *Account) Login() (*client.Client, error) {
	c, err := client.DialTLS(acc.Host, nil)
	c.SetDebug(os.Stderr)
	if err != nil {
		return nil, err
	}
	if err := c.Login(acc.User, acc.Pass); err != nil {
		return nil, err
	}
	return c, nil
}

// MailBoxUpdate ...
type MailBoxUpdate struct {
	Update      client.Update
	MboxID      int
	NeedRefresh bool
}

// MailBoxConn ...
type MailBoxConn struct {
	Mbox       *MailBox
	client     *client.Client
	idleClient *idle.Client
	stop       chan struct{}
}

// StopIdle ...
func (conn *MailBoxConn) StopIdle() {
	conn.stop <- struct{}{}
}

// Idle ...
func (conn *MailBoxConn) Idle(mboxID int, globalUpdates chan MailBoxUpdate) error {
	defer conn.client.Logout()
	updates := make(chan client.Update)
	conn.client.Updates = updates
	done := make(chan error, 1)
	go func() {
		done <- conn.idleClient.Idle(conn.stop)
	}()
	for {
		select {
		case update := <-updates:
			globalUpdates <- MailBoxUpdate{
				Update:      update,
				MboxID:      mboxID,
				NeedRefresh: false,
			}

		case err := <-done:
			conn.client.Updates = nil
			if err != nil {
				log.Println("WUT???", err)
				globalUpdates <- MailBoxUpdate{
					Update:      nil,
					MboxID:      mboxID,
					NeedRefresh: true,
				}
				return err
			}
			return nil
		}
	}
}

// MailBox ...
type MailBox struct {
	Account *Account
	MailBox string
}

// Connect ...
func (mbox *MailBox) Connect() (*MailBoxConn, error) {
	c, err := mbox.Account.Login()
	if err != nil {
		return nil, err
	}
	if _, err := c.Select(mbox.MailBox, false); err != nil {
		return nil, err
	}
	idleClient := idle.NewClient(c)
	if ok, err := idleClient.SupportIdle(); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("Host %v doesn't support IMAP IDLE", mbox.Account.Host)
	}
	return &MailBoxConn{
		client:     c,
		idleClient: idleClient,
		Mbox:       mbox,
		stop:       make(chan struct{}),
	}, nil
}
