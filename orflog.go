package orflog

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/go-pkgz/lgr"
)

// Service create engine to collects logs from orf
type Service struct {
	Opts

	sync.WaitGroup

	logMapAll struct {
		sync.RWMutex
		m map[string]Orf
	}

	logMapOld struct {
		sync.RWMutex
		m map[string]Orf
	}

	newLogCh    chan Orf
	removeLogCh chan Orf

	timeStart time.Time
}

// Opts collects parameters to initialize Service
type Opts struct {
	LogPaths  []string      `long:"log-paths" env:"LOG_PATHS" description:"path to log files" env-delim:","`
	LogSuffix string        `long:"log-suffix" env:"LOG_SUFFIX" default:".log" description:"log file extension"`
	OrfLine   string        `long:"orfline" env:"ORFLINE" default:"SMTPSVC" description:"search start word in log line"`
	SleepTime time.Duration `long:"sleep-time" env:"SLEEP_TIME" default:"1m" description:"sleep time after every run"`
	TimeRange struct {
		Years  int `long:"years" env:"YEARS" default:"0" description:"years time range for logs"`
		Months int `long:"months" env:"MONTHS" default:"1" description:"months time range for logs"`
		Days   int `long:"days" env:"DAYS" default:"0" description:"days time range for logs"`
	} `group:"time-range" namespace:"time-range" env-namespace:"TIME_RANGE"`
}

const (
	logSuffix = ".log"
	orfLine   = "SMTPSVC"
	sleepTime = 10 * time.Second
)

// NewService initialize everything
func NewService(opts Opts) *Service {
	res := &Service{
		Opts: opts,
	}

	if res.LogSuffix == "" {
		res.LogSuffix = logSuffix
	}

	if res.OrfLine == "" {
		res.OrfLine = orfLine
	}

	if res.TimeRange.Years == 0 && res.TimeRange.Months == 0 && res.TimeRange.Days == 0 {
		res.TimeRange.Months = 1
	}

	if res.SleepTime.Seconds() < 1 {
		res.SleepTime = sleepTime
	}

	res.newLogCh = make(chan Orf)
	res.removeLogCh = make(chan Orf)
	res.logMapAll.m = make(map[string]Orf)
	res.logMapOld.m = make(map[string]Orf)

	return res
}

// Run service loop
func (s *Service) Run(ctx context.Context) {
	s.timeStart = time.Now().AddDate(-s.TimeRange.Years, -s.TimeRange.Months, -s.TimeRange.Days)

	for {
		select {
		case <-ctx.Done():
			<-ctx.Done()
			log.Printf("[WARN] init terminate service")
			s.closeChannels()
			log.Printf("[WARN] service terminated")
			return
		default:
			logFiles := s.getLastModifiedLogFiles()
			log.Printf("logFiles: %d", len(logFiles))

			allStrings := s.getAllStringsFromLogFiles(logFiles)
			log.Printf("allStrings: %d", len(allStrings))

			s.logMapOld.m = make(map[string]Orf)

			orfs := s.createOrfRecords(allStrings)
			log.Printf("orfs: %d", len(orfs))

			s.removeOldRecords()

			for _, orf := range orfs {
				s.newLogCh <- orf
			}

			s.timeStart = time.Now().Add(-10 * time.Minute)
			time.Sleep(s.SleepTime)
		}
	}
}

// GetLastRecords from last program start
func (s *Service) GetLastRecords() []Orf {
	logFiles := s.getLastModifiedLogFiles()
	log.Printf("logFiles: %d", len(logFiles))

	allStrings := s.getAllStringsFromLogFiles(logFiles)
	log.Printf("allStrings: %d", len(allStrings))

	s.logMapOld.m = make(map[string]Orf)

	orfs := s.createOrfRecords(allStrings)
	log.Printf("orfs: %d", len(orfs))

	s.removeOldRecords()

	s.timeStart = time.Now().Add(-10 * time.Minute)
	return orfs
}

// Channel return channels with new and old records
func (s *Service) Channel() (new <-chan Orf) {
	return s.newLogCh
}

func (s *Service) closeChannels() {
	close(s.newLogCh)
	close(s.removeLogCh)
}

func (s *Service) getLastModifiedLogFiles() []string {
	result := make([]string, 0)
	for _, dir := range s.LogPaths {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Printf("[WARN] could not open directory %s: %v", dir, err)
			continue
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), s.LogSuffix) && file.ModTime().After(s.timeStart) {
				fileName := file.Name()
				result = append(result, filepath.Join(dir, fileName))
			}
		}
	}

	return result
}

func (s *Service) getAllStringsFromLogFiles(fileNames []string) []string {
	result := make([]string, 0)

	for _, fileName := range fileNames {
		b, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Printf("[WARN] could not read file %s: %v", fileName, err)
			continue
		}

		lines := strings.Split(string(b), "\n")

		for _, line := range lines {
			result = append(result, line)
		}
	}

	return result
}

func (s *Service) createOrfRecords(lines []string) []Orf {
	result := make([]Orf, 0)
	for _, line := range lines {
		if strings.Contains(line, s.OrfLine) {
			splitString := strings.Split(line, " ")

			messageFromSplit := splitString[12:]

			var message bytes.Buffer
			if len(messageFromSplit) != 0 {
				for _, msg := range messageFromSplit {
					message.WriteString(msg + " ")
				}
			}

			const timeFormat = "2006-01-02T15:04:05"
			timeFromSplit := splitString[1]
			t, err := time.Parse(timeFormat, timeFromSplit)
			if err != nil {
				log.Printf("[WARN] could not parse time %v: %v", timeFromSplit, err)
				continue
			}

			orf := Orf{
				Time:           t,
				Action:         ifReject(splitString[4]),
				FilteringPoint: filterPoint(splitString[5]),
				RelatedIP:      splitString[6],
				Sender:         splitString[7],
				Recipients:     splitString[8],
				Message:        message.String(),
			}

			if strings.Contains(orf.Recipients, ";") {
				splitRecipients := strings.Split(orf.Recipients, ";")

				for _, recipient := range splitRecipients {
					orf.Recipients = recipient

					s.appendRecord(orf, &result)
				}
			} else {
				s.appendRecord(orf, &result)
			}
		}
	}

	return result
}

func (s *Service) appendRecord(orf Orf, result *[]Orf) {
	orf.Hash()
	if err := s.addRecordToMaps(orf); err != nil {
		return
	}
	*result = append(*result, orf)
}

func (s *Service) addRecordToMaps(orf Orf) error {
	if orf.Time.Before(s.timeStart) {
		s.logMapOld.Lock()
		s.logMapOld.m[orf.HashString] = orf
		s.logMapOld.Unlock()
		return errors.New("record time before needed time")
	}

	s.logMapAll.RLock()
	_, ok := s.logMapAll.m[orf.HashString]
	s.logMapAll.RUnlock()

	if ok {
		return errors.New("this record already in chan")
	}

	s.logMapAll.Lock()
	s.logMapAll.m[orf.HashString] = orf
	s.logMapAll.Unlock()

	return nil
}

func (s *Service) removeOldRecords() {
	for h := range s.logMapOld.m {
		if _, ok := s.logMapAll.m[h]; ok {
			s.logMapAll.Lock()
			delete(s.logMapAll.m, h)
			s.logMapAll.Unlock()
		}
	}
}

func ifReject(s string) string {
	switch s {
	case "Reject":
		return "Не доставлено"
	case "RemoveRecipient":
		return "Не доставлено из-за отсутствия получателя"
	case "ReplaceAttachment":
		return "Доставлено с удалением вложения"
	case "WhitelistRecipient":
		return "Доставлено принудительно"
	default:
		return "Доставлено"
	}
}

func filterPoint(s string) string {
	switch s {
	case "BeforeArrival":
		return "Отфильтровано до прибытия"
	case "OnArrival":
		return "Отфильтровано во время прибытия"
	default:
		return s
	}
}
