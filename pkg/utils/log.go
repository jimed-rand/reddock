package utils

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"redway/pkg/config"
)

type LogManager struct {
	config *config.Config
}

func NewLogManager() *LogManager {
	cfg, _ := config.Load()
	return &LogManager{config: cfg}
}

func (l *LogManager) Show() error {
	logFile := l.config.LogFile

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return fmt.Errorf("The log file '%s' is not found", logFile)
	}

	fmt.Printf("Showing logs from: %s\n", logFile)
	fmt.Println("Press Ctrl+C to exit")

	cmd := exec.Command("tail", "-f", logFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cmd.Process.Signal(syscall.SIGTERM)
	}()

	return cmd.Wait()
}
