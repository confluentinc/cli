package config

import (
	"github.com/fatih/color"

	"github.com/confluentinc/go-prompt"
)

// Package used as a root for color used across the application
const PromptAccentColor = prompt.Cyan
const AccentColor = color.FgCyan

const InfoColor = color.FgWhite
const ErrorColor = color.FgHiRed
const WarnColor = color.FgCyan
