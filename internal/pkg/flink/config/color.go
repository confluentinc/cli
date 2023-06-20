package config

import (
	"github.com/confluentinc/go-prompt"
	"github.com/fatih/color"
)

// Package used as a root for color used accross the application
const PromptAccentColor = prompt.Cyan
const AccentColor = color.FgCyan

const InfoColor = color.FgWhite
const ErrorColor = color.FgHiRed
const WarnColor = color.FgHiYellow
