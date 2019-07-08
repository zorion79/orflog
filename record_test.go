package orflog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrf_Hash(t *testing.T) {
	orf := Orf{
		Action: "action",
	}

	orf.Hash()

	assert.Equal(t, "9f9a426dae97fe1ed2ade244d2c246d7", orf.HashString)
}
