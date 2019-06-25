package orflog

import (
	"crypto/md5"
	"encoding/json"
	"time"
)

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

// OrfHash return hash of Orf
func (o *Orf) OrfHash() [16]byte {
	jsonBytes, _ := json.Marshal(o)
	return md5.Sum(jsonBytes)
}
