package main

import (
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"

	"github.com/cnt0/maildir_idle_sync/config"
	"github.com/cnt0/maildir_idle_sync/manager"
)

func main() {

	var cfg config.Config
	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		panic(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	mgr, err := manager.NewConnectionManager(&cfg)
	if err != nil {
		panic(err)
	}
	mgr.Idle(interrupt)
}
