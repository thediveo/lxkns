package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	logrus "github.com/sirupsen/logrus"
	"github.com/thediveo/gons/reexec"
	"github.com/thediveo/lxkns"
	log "github.com/thediveo/lxkns/log"
	_ "github.com/thediveo/lxkns/log/logrus"
)

func main() {
	// For some discovery methods this app must be forked and re-executed; the
	// call to reexec.CheckAction() will automatically handle this situation
	// and then never return when in re-execution.
	reexec.CheckAction()

	// FIXME: use unified interface?
	logrus.SetLevel(logrus.DebugLevel)
	log.SetLevel(log.DebugLevel)

	// And now for the real meat.
	log.Infof("this is the lxkns Linux-kernel namespaces discovery service version %s", lxkns.SemVersion)
	log.Infof("https://github.com/thediveo/lxkns")
	if _, err := startServer("[::]:5010"); err != nil {
		log.Errorf("cannot start service, error: %s", err.Error())
		os.Exit(1)
	}
	stopit := make(chan os.Signal, 1)
	signal.Notify(stopit, syscall.SIGINT)
	signal.Notify(stopit, syscall.SIGTERM)
	signal.Notify(stopit, syscall.SIGQUIT)
	<-stopit
	stopServer(15 * time.Second)
}
