package types

import (
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
