//go:build !windows
// +build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/CloudPassenger/rnm-go/config"
)

func signalHandler(oldConf *config.Config) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)
	for range ch {
		ReloadConfig(oldConf)
	}
}
