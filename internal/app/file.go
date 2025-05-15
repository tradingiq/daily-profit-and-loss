package app

import (
	"os"
	"path/filepath"
)

func GetDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".daily-pnl"
	}
	return filepath.Join(homeDir, ".daily-pnl")
}

func GetLogDirectory() string {
	return filepath.Join(GetDirectory(), "logs")
}

func GetConfigPath() string {
	appDir := GetDirectory()

	return filepath.Join(appDir, "config.json")
}
