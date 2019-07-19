package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/go-pkgz/lgr"
	"github.com/zorion79/orflog/v3"
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
		TimeRange: struct {
			Years  int `long:"years" env:"YEARS" default:"0" description:"years time range for logs"`
			Months int `long:"months" env:"MONTHS" default:"1" description:"months time range for logs"`
			Days   int `long:"days" env:"DAYS" default:"0" description:"days time range for logs"`
		}{
			Years:  0,
			Months: 1,
			Days:   0,
		},
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

	orfs := service.GetLastRecords()
	log.Printf("Orfs = %d", len(orfs))

	go service.Run(ctx)

	newLogCh := service.Channel()

	for orf := range newLogCh {
		log.Printf("%+v", orf)
	}
}
