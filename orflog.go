package orflog

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Service create engine to collects logs from orf
type Service struct {
	Opts
}

// Opts collects parameters to initialize Service
type Opts struct {
	LogPaths                        []string // directory path `\\orf01\ORF\`
	LogSuffix                       string   // .log
	GetAllStringsGoroutinesCount    int      // 40 is ok
	CreateOrfRecordsGoroutinesCount int      // 10 000 is ok
	OrfLine                         string   // SMTPSVC
	TimeRange                       struct {
		Years  int
		Months int
		Days   int
	}
}

// NewService initialize everything
func NewService(opts Opts) *Service {
	return &Service{
		opts,
	}
}

// GetLogs return orf logs
func (s *Service) GetLogs() (res []Orf) {
	orfLogFiles := s.getLastModifiedLogFiles()
	stringsChan := s.getAllStringsFromLogFiles(orfLogFiles)
	recordsChan := s.createOrfRecords(stringsChan)

	for orf := range recordsChan {
		res = append(res, orf)
	}

	return res
}

func (s *Service) getLastModifiedLogFiles() <-chan string {
	result := make(chan string)

	t := time.Now().AddDate(-s.TimeRange.Years, -s.TimeRange.Months, -s.TimeRange.Days)

	var wg sync.WaitGroup
	wg.Add(len(s.LogPaths))

	for _, dir := range s.LogPaths {
		go func(dir string) {
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				log.Printf("[WARN] could not take files from directory %s: %v", dir, err)
				wg.Done()
				return
			}

			fileName := ""
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), s.LogSuffix) && file.ModTime().After(t) {
					fileName = file.Name()
					result <- filepath.Join(dir, fileName)
				}
			}
			wg.Done()
		}(dir)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

func (s *Service) getAllStringsFromLogFiles(fileNames <-chan string) <-chan string {
	result := make(chan string)

	var wg sync.WaitGroup
	for i := 0; i < s.GetAllStringsGoroutinesCount; i++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case fileName, ok := <-fileNames:
					if !ok {
						wg.Done()
						return
					}
					b, err := ioutil.ReadFile(fileName)
					if err != nil {
						log.Printf("[WARN] could not read file %s: %v", fileName, err)
						return
					}

					lines := strings.Split(string(b), "\n")

					for _, line := range lines {
						result <- line
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

func (s *Service) createOrfRecords(stringsChan <-chan string) <-chan Orf {
	result := make(chan Orf)

	var wg sync.WaitGroup

	for i := 0; i < s.CreateOrfRecordsGoroutinesCount; i++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case line, ok := <-stringsChan:
					if !ok {
						wg.Done()
						return
					}
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
							log.Printf("[WARN] could not parce time %v: %v", timeFromSplit, err)
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
						orf.HashString = fmt.Sprintf("%x", orfHash(orf))

						if strings.Contains(orf.Recipients, ";") {
							splitRecipients := strings.Split(orf.Recipients, ";")

							for _, recipient := range splitRecipients {
								orf.Recipients = recipient
								result <- orf
							}
						} else {
							result <- orf
						}
					}
				}
			}
		}()
	}
	return result
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

func orfHash(o Orf) [16]byte {
	jsonBytes, _ := json.Marshal(o)
	return md5.Sum(jsonBytes)
}
