IMAP IDLE client.  
WIP, seems to be working but not ready for production.  

`config.toml` - config example.  
Pass it with `-c` command option. `maildir_idle_sync -c /path/to/config.toml`  
`xxx.UpdateEvents` blocks are not mandatory, you can skip any one of them.  
If there will be no suitable command in `[Account.Mailboxes.UpdateEvents]`, it will try to find the one in `[Account.UpdateEvents]`, and then in `[UpdateEvents]`.  
If `PasswordCommand=` is not empty, then password from `Pass=` will be replaced with its output. Both entries are optional.    

`maildir_idle.service` - systemd service example.  
Notice `KillSignal=SIGINT` - this is mandatory for graceful shutdown.