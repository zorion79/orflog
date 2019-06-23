package orflog

import "time"

//Orf collects fields from orf log file
type Orf struct {
	Time           time.Time
	Action         string
	FilteringPoint string
	RelatedIP      string
	Sender         string
	Recipients     string
	Message        string
	HashString     string
}
