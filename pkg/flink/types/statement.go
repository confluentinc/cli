package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func GenerateStatementName() string {
	clientName := "cli"
	date := time.Now().Format("2006-01-02")
	localTime := time.Now().Format("150405")
	id := uuid.New().String()
	return fmt.Sprintf("%s-%s-%s-%s", clientName, date, localTime, id)
}

func GenerateStatementNameForOnPrem() string {
	clientName := "cli"
	date := time.Now().Format("20060102")    // 8 chars
	localTime := time.Now().Format("150405") // 6 chars
	// 12 random bytes => 24 hex chars
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	randomHex := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s", clientName, date, localTime, randomHex)
}
