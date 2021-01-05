package cennzTransaction

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestGetEra(t *testing.T) {
	height := uint64(0)

	era := GetEra(height)

	fmt.Println(hex.EncodeToString(era))
}
