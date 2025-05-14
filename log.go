package main

import (
	"os/exec"
	"runtime"
)

func openLogFile() {
	logFilePath := log.GetLogFilePath()
	log.Info("opening log file: %s", logFilePath)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("notepad.exe", logFilePath)
	case "darwin":
		cmd = exec.Command("open", logFilePath)
	default:
		cmd = exec.Command("xdg-open", logFilePath)
	}

	if err := cmd.Start(); err != nil {
		log.Error("Failed to open log file: %v", err)
	}
}
