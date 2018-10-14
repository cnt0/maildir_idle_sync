package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/BurntSushi/toml"

	"github.com/cnt0/maildir_idle_sync/config"
	utils "github.com/cnt0/maildir_idle_sync/conn-utils"
	"github.com/cnt0/maildir_idle_sync/imap"
	events "github.com/cnt0/maildir_idle_sync/update-events"
)

// EventsWithAccID ...
type EventsWithAccID struct {
	ev    *events.UpdateEvents
	accID int
}

func kek(a, b *int) {
	if a == nil {
		a = new(int)
	}
	*a = *b + 1
	fmt.Println("a", *a)
}

func main() {

	// var a *int
	// b := 55
	// kek(a, &b)
	// fmt.Println(a)
	// return

	var cfg config.Config
	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		log.Fatal(err)
	}

	// if err := cfg.UpdateEvents.Default.Run(); err != nil {
	// 	log.Fatal(err)
	// }
	// return

	mboxes := []*imap.MailBox{}
	accEvents := []*events.UpdateEvents{}
	mboxEvents := []EventsWithAccID{}

	for accID, acc := range cfg.Account {
		accEvents = append(accEvents, acc.UpdateEvents)
		for _, box := range acc.Mailboxes {
			mboxes = append(mboxes, &imap.MailBox{
				Account: acc.Account,
				MailBox: box.Name,
			})

			mboxEvents = append(mboxEvents, EventsWithAccID{
				ev:    box.UpdateEvents,
				accID: accID,
			})
		}
	}

	log.Println("START")
	interrupt := make(chan os.Signal, 1)

	signal.Notify(interrupt, os.Interrupt)

	var wg sync.WaitGroup
	updates := make(chan imap.MailBoxUpdate)
	conns := make([]*imap.MailBoxConn, len(mboxes))

	if err := utils.RefreshConns(conns, mboxes, updates, &wg); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case u := <-updates:
			if u.NeedRefresh {
				log.Printf("mbox %v NEEDS REFRESH!", mboxes[u.MboxID].MailBox)
				if err := utils.RefreshConn(u.MboxID, &wg, conns[u.MboxID], mboxes[u.MboxID], updates); err != nil {
					log.Fatal(err)
				}
				break
			}
			if err := mboxEvents[u.MboxID].ev.Handle(u.Update); err != events.ErrNoEventHandler {
				if err != nil {
					log.Println(err)
				}
				break
			}
			if err := accEvents[mboxEvents[u.MboxID].accID].Handle(u.Update); err != events.ErrNoEventHandler {
				if err != nil {
					log.Println(err)
				}
				break
			}
			if err := cfg.UpdateEvents.Handle(u.Update); err != nil {
				log.Println(err)
			}
		case <-interrupt:
			utils.CloseConns(conns, &wg)
			return
		}
	}
}
