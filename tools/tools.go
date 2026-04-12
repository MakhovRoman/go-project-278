package tools

import (
	"crypto/rand"
	"log"
)
import "encoding/hex"

func GenerateShortName() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("failed to generate shortname: %v", err)
		return ""
	}
	return hex.EncodeToString(b)
}
