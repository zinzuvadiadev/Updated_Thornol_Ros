package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type CmdLogger struct {
	stdout   io.WriteCloser
	stderr   io.WriteCloser
	combined io.WriteCloser
}

func NewCmdLogger(processName string) (*CmdLogger, error) {
	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create timestamp for log files
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	
	// Create log files with timestamp
	stdoutPath := filepath.Join(logsDir, fmt.Sprintf("%s_%s_stdout.log", processName, timestamp))
	stderrPath := filepath.Join(logsDir, fmt.Sprintf("%s_%s_stderr.log", processName, timestamp))
	combinedPath := filepath.Join(logsDir, fmt.Sprintf("%s_%s_combined.log", processName, timestamp))

	stdout, err := os.Create(stdoutPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout log file: %w", err)
	}

	stderr, err := os.Create(stderrPath)
	if err != nil {
		stdout.Close()
		return nil, fmt.Errorf("failed to create stderr log file: %w", err)
	}

	combined, err := os.Create(combinedPath)
	if err != nil {
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("failed to create combined log file: %w", err)
	}

	return &CmdLogger{
		stdout:   stdout,
		stderr:   stderr,
		combined: combined,
	}, nil
}

func (l *CmdLogger) Close() {
	if l.stdout != nil {
		l.stdout.Close()
	}
	if l.stderr != nil {
		l.stderr.Close()
	}
	if l.combined != nil {
		l.combined.Close()
	}
}

func (l *CmdLogger) GetWriters() (io.Writer, io.Writer) {
	stdout := io.MultiWriter(l.stdout, l.combined)
	stderr := io.MultiWriter(l.stderr, l.combined)
	return stdout, stderr
} 