package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"

	"github.com/cnt0/maildir_idle_sync/config"
	"github.com/cnt0/maildir_idle_sync/manager"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "c", "config.toml", "path to toml config file")
}

func main() {

	flag.Parse()

	var cfg config.Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		panic(err)
	}

	for i := range cfg.Account {
		if len(cfg.Account[i].PasswordCommand) > 0 {
			cfg.Account[i].Pass = string(cfg.Account[i].PasswordCommand)
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	mgr, err := manager.NewConnectionManager(&cfg)
	if err != nil {
		panic(err)
	}
	mgr.Idle(interrupt)
}
