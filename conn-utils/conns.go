package utils

import (
	"fmt"
	"log"
	"sync"

	"github.com/cnt0/maildir_idle_sync/imap"
)

// CloseConns ...
func CloseConns(conns []*imap.MailBoxConn, wg *sync.WaitGroup) {
	for _, conn := range conns {
		if conn != nil {
			conn.StopIdle()
		}
	}
	wg.Wait()
}

// RefreshConn ...
func RefreshConn(
	idx int,
	wg *sync.WaitGroup,
	conn *imap.MailBoxConn,
	mbox *imap.MailBox,
	updates chan imap.MailBoxUpdate,
) error {
	wg.Add(1)
	conn1, err := mbox.Connect()
	if err != nil {
		return fmt.Errorf(
			"error connecting to %v (%v): %v",
			mbox.Account.Name,
			mbox.MailBox,
			err,
		)
	}
	*conn = *conn1
	go func(idx int, account, mbox string, wg *sync.WaitGroup) {
		defer wg.Done()
		log.Printf("%v: START IDLING", mbox)
		if err := conn.Idle(idx, updates); err != nil {
			log.Printf("Error in %v (%v): %v\n", account, mbox, err)
		}
	}(idx, mbox.Account.Name, mbox.MailBox, wg)
	return nil
}

// RefreshConns ...
func RefreshConns(
	conns []*imap.MailBoxConn,
	mboxes []*imap.MailBox,
	updates chan imap.MailBoxUpdate,
	wg *sync.WaitGroup,
) error {

	CloseConns(conns, wg)

	for i, mbox := range mboxes {
		conns[i] = new(imap.MailBoxConn)
		if err := RefreshConn(i, wg, conns[i], mbox, updates); err != nil {
			return err
		}
	}
	return nil
}
