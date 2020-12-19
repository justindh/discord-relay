package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/justindh/discord-relay/internal/config"
	"github.com/justindh/discord-relay/internal/forwarder"
	"github.com/justindh/discord-relay/internal/listener"
)

func init() {
	var debug = flag.Bool("v", true, "enables verbose logging")
	// Setup logging options
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetReportCaller(true)

}

func main() {
	var configPath = flag.String("c", "config.yaml", "specicies the path to the config")

	// Parse up the command line flags
	flag.Parse()

	// Make sure the config exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Errorf("config doesnt exist: %s", err)
		return
	}

	//Load Config
	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Errorf("Error getting config: %s", err)
		return
	}

	// Connect to the destination for all messages
	f, err := forwarder.NewForwarder(cfg.ForwarderToken, cfg.IsWebhook, cfg.ChannelMap())
	if err != nil {
		log.Errorf("Error while creating Forwarder: %s", err)
		return
	}
	defer f.Close()

	// Test that we can send to this correctly
	err = f.Send("[log] Forwarder Connected", cfg.ErrorLogChannelID)
	if err != nil {
		log.Errorf("Error while sending to log: %s", err)
		return
	}

	// Open up all the listners and start processing messages.
	l, err := listener.NewListeners(cfg.ListenerTokens(), f)
	if err != nil {
		log.Errorf("Error while creating listeners: %s", err)
		return
	}

	defer l.Close()

	log.Infoln("Relay is now running.  Press CTRL-C to exit.")

	//TODO this isnt the right way to do this....
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Infoln("Shutting down...")

}
