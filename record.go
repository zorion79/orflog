package orflog

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
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

// Hash return hash of Orf
func (o *Orf) Hash() {
	jsonBytes, _ := json.Marshal(o)
	o.HashString = fmt.Sprintf("%x", md5.Sum(jsonBytes))
}
