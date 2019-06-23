package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

	getAllStringsGoroutinesCount, err := strconv.Atoi(os.Getenv("EXMPL_ORFLOG_GOROUTINECOUNT_STRINGS"))
	if err != nil {
		log.Fatalf("could not get env: %v", err)
	}

	createOrfRecordsGoroutinesCount, err := strconv.Atoi(os.Getenv("EXMPL_ORFLOG_GOROUTINECOUNT_ORFRECORDS"))
	if err != nil {
		log.Fatalf("could not get env: %v", err)
	}

	years, err := strconv.Atoi(os.Getenv("EXMPL_ORFLOG_YEARS"))
	if err != nil {
		log.Fatalf("could not get env: %v", err)
	}

	months, err := strconv.Atoi(os.Getenv("EXMPL_ORFLOG_MOTHS"))
	if err != nil {
		log.Fatalf("could not get env: %v", err)
	}

	days, err := strconv.Atoi(os.Getenv("EXMPL_ORFLOG_DAYS"))
	if err != nil {
		log.Fatalf("could not get env: %v", err)
	}

	options := orflog.Opts{
		LogPaths:                        logPaths,
		LogSuffix:                       os.Getenv("EXMPL_ORFLOG_LOGSUFFIX"),
		GetAllStringsGoroutinesCount:    getAllStringsGoroutinesCount,
		CreateOrfRecordsGoroutinesCount: createOrfRecordsGoroutinesCount,
		OrfLine:                         os.Getenv("EXMPL_ORFLOG_ORFLINE"),
		TimeRange: struct {
			Years  int
			Months int
			Days   int
		}{
			Years:  years,
			Months: months,
			Days:   days,
		},
	}

	service := orflog.NewService(options)

	logs := service.GetLogs()
	fmt.Printf("len of logs=%d", len(logs))
}
