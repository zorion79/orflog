package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/go-pkgz/lgr"
	"github.com/zorion79/orflog"
)

func main() {
	log.Setup(log.Debug, log.Msec, log.LevelBraces, log.CallerFile, log.CallerFunc) // setup default logger with go-pkgz/lgr

	logPathFromEnv := os.Getenv("EXMPL_ORFLOG_LOGPATH") // \\orf01\ORF\,\\orf02\ORF\
	logPaths := []string{logPathFromEnv}
	if strings.Contains(logPathFromEnv, ",") {
		logPaths = strings.Split(logPathFromEnv, ",")
	}

	options := orflog.Opts{
		LogPaths: logPaths,
	}

	log.Printf("[DEBUG] opts=%+v", options)

	service := orflog.NewService(options)
	ctx, cancel := context.WithCancel(context.Background())
	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()
	go service.Run(ctx)

	newLogCh, removeLogCh := service.Channel()

	for {
		select {
		case newRecord, ok := <-newLogCh:
			if !ok {
				log.Printf("[WARN] program closed")
				return
			}
			log.Printf("new %+v", newRecord)

		case oldRecord, ok := <-removeLogCh:
			if !ok {
				log.Printf("[WARN] program closed")
			}
			log.Printf("old %+v", oldRecord)
		}
	}
}
