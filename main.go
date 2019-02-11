package main

import (
	"os"

	c "github.com/helto4real/go-daemon/daemon/core"
	"github.com/helto4real/go-hassclient/client"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log *logrus.Entry

func main() {

	log.Println("Starting go-daemon...")

	osSignal := make(chan os.Signal, 1)
	daemon := c.NewApplicationDaemon()
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
func init() {
	log = logrus.WithField("prefix", "go-appdaemon")
	Formatter := new(prefixed.TextFormatter)
	Formatter.FullTimestamp = true
	Formatter.TimestampFormat = "2006-01-02 15:04:05"
	Formatter.DisableColors = true
	Formatter.ForceColors = false
	Formatter.ForceFormatting = true
	logrus.SetFormatter(Formatter)
	logrus.SetLevel(logrus.InfoLevel)

}
