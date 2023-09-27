package flink

import "github.com/google/uuid"

func GenerateStatementName() string {
	return uuid.New().String()[:18]
}

const ServiceAccountWarning = "[WARN] No service account provided. To ensure that your statements run continuously, " +
	"use a service account instead of your user identity with `confluent iam service-account use` or `--service-account`. " +
	"Otherwise, statements will stop running after 4 hours."
