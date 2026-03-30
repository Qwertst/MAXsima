package main

import (
	"fmt"
	"os"

	"github.com/aydreq/maxsima/config"
	"github.com/aydreq/maxsima/internal/chat"
	"github.com/aydreq/maxsima/internal/client"
	"github.com/aydreq/maxsima/internal/server"
	"github.com/aydreq/maxsima/internal/ui"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	console := ui.NewConsoleUI(os.Stdout, os.Stdin)
	manager := chat.NewManager(cfg.Username, console)

	if cfg.IsServerMode() {
		srv := server.New(manager)
		if err := server.Listen(cfg.Port, srv); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		cli, err := client.New(cfg.PeerAddress, manager)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := cli.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
