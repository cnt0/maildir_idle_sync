package manager

import (
	"fmt"
	"os"
	"sync"

	"github.com/cnt0/maildir_idle_sync/config"
	"github.com/cnt0/maildir_idle_sync/imap"
	events "github.com/cnt0/maildir_idle_sync/update-events"
)

// ConnectionManager manages IDLE connections to IMAP mailboxes
type ConnectionManager struct {
	wg sync.WaitGroup

	updates chan imap.MailBoxUpdate
	mboxes  []*imap.MailBox
	conns   []*imap.MailBoxConn

	globalEvents *events.UpdateEvents
	accEvents    []*events.UpdateEvents
	mboxEvents   []eventsWithAccID
}

type eventsWithAccID struct {
	ev    *events.UpdateEvents
	accID int
}

// NewConnectionManager creates ConnectionManager from given config
func NewConnectionManager(cfg *config.Config) (*ConnectionManager, error) {

	mboxes := []*imap.MailBox{}
	accEvents := []*events.UpdateEvents{}
	mboxEvents := []eventsWithAccID{}

	for accID, acc := range cfg.Account {
		accEvents = append(accEvents, acc.UpdateEvents)
		for _, box := range acc.Mailboxes {
			mboxes = append(mboxes, &imap.MailBox{
				Account: acc.Account,
				MailBox: box.Name,
			})

			mboxEvents = append(mboxEvents, eventsWithAccID{
				ev:    box.UpdateEvents,
				accID: accID,
			})
		}
	}

	mgr := &ConnectionManager{
		wg: sync.WaitGroup{},

		updates: make(chan imap.MailBoxUpdate),
		mboxes:  mboxes,
		conns:   make([]*imap.MailBoxConn, len(mboxes)),

		globalEvents: cfg.UpdateEvents,
		accEvents:    accEvents,
		mboxEvents:   mboxEvents,
	}

	if err := mgr.refreshConns(); err != nil {
		return nil, err
	}

	return mgr, nil
}

func (mgr *ConnectionManager) refreshConn(idx int) error {
	mgr.wg.Add(1)
	conn, err := mgr.mboxes[idx].Connect()
	if err != nil {
		return fmt.Errorf(
			"error connecting to %v (%v): %v",
			mgr.mboxes[idx].Account.Name,
			mgr.mboxes[idx].MailBox,
			err,
		)
	}
	mgr.conns[idx] = conn
	go func(account, mbox string) {
		defer mgr.wg.Done()
		if err := conn.Idle(idx, mgr.updates); err != nil {
			fmt.Printf("error in %v (%v): %v", account, mbox, err)
		}
	}(mgr.mboxes[idx].Account.Name, mgr.mboxes[idx].MailBox)
	return nil
}

func (mgr *ConnectionManager) closeConns() {
	for _, conn := range mgr.conns {
		if conn != nil {
			conn.StopIdle()
		}
	}
	mgr.wg.Wait()
}

func (mgr *ConnectionManager) refreshConns() error {
	mgr.closeConns()
	for i := range mgr.conns {
		err := mgr.refreshConn(i)
		if err != nil {
			return err
		}
	}
	return nil
}

// Idle processes updates from mailbox connections.
// Send signal to interrupt to, you guess, interrupt it
func (mgr *ConnectionManager) Idle(interrupt chan os.Signal) {
	for {
		select {
		case u := <-mgr.updates:
			if u.NeedRefresh {
				fmt.Printf("mbox %v needs refresh", mgr.mboxes[u.MboxID].MailBox)
				err := mgr.refreshConn(u.MboxID)
				if err != nil {
					panic(err)
				}
				break
			}
			if mgr.mboxEvents[u.MboxID].ev != nil {
				fmt.Printf("%v: using mbox event", mgr.mboxes[u.MboxID].Account.Name)
				err := mgr.mboxEvents[u.MboxID].ev.Handle(u.Update)
				if err != events.ErrNoEventHandler {
					if err != nil {
						fmt.Println(err)
					}
					break
				}
			}
			if mgr.accEvents[mgr.mboxEvents[u.MboxID].accID] != nil {
				fmt.Printf("%v: using account event", mgr.mboxes[u.MboxID].Account.Name)
				err := mgr.accEvents[mgr.mboxEvents[u.MboxID].accID].Handle(u.Update)
				if err != events.ErrNoEventHandler {
					if err != nil {
						fmt.Println(err)
					}
					break
				}
			}
			if mgr.globalEvents != nil {
				fmt.Printf("%v: using global event", mgr.mboxes[u.MboxID].Account.Name)
				if err := mgr.globalEvents.Handle(u.Update); err != nil {
					fmt.Println(err)
				}
			}
		case <-interrupt:
			mgr.closeConns()
			return
		}
	}
}
