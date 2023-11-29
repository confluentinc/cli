package types

import "github.com/google/uuid"

func GenerateStatementName() string {
	return uuid.New().String()[:18]
}
