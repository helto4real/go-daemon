package main

import (
	"log"
	"os"

	"github.com/helto4real/go-daemon/daemon"
	"github.com/helto4real/go-hassclient/client"
)

func main() {
	osSignal := make(chan os.Signal, 1)
	daemon := daemon.NewApplicationDaemon()
	hass := client.NewHassClient()
	// Apps is defined in the apps.go file
	daemon.Start(".", hass, apps)

	for {
		select {
		case <-osSignal:
			log.Println("OS SIGNAL")
			daemon.Stop()
		}
	}
}
