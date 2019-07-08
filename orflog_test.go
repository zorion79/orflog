package orflog

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_NewService(t *testing.T) {
	opt := Opts{
		LogPaths: []string{"./testf", "./test"},
	}

	svc := NewService(opt)
	assert.Equal(t, ".log", svc.LogSuffix)
	assert.Equal(t, "SMTPSVC", svc.OrfLine)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go svc.Run(ctx)

	newOrfs, _ := svc.Channel()
	for newOrf := range newOrfs {
		t.Logf("Orf: %+v", newOrf)
	}
}
