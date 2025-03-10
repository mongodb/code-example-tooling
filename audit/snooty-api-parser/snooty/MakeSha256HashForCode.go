package snooty

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func MakeSha256HashForCode(code string) string {
	whiteSpaceTrimmedNode := strings.TrimSpace(code)
	hash := sha256.Sum256([]byte(whiteSpaceTrimmedNode))
	return hex.EncodeToString(hash[:])
}
