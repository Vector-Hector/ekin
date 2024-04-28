package main

import (
	"encoding/csv"
	"os"
	"sync"
)

// CsvLogger struct to manage CSV file logging
type CsvLogger struct {
	file   *os.File
	writer *csv.Writer
	mutex  sync.Mutex
}

// NewCsvLogger initializes and returns a new CsvLogger
func NewCsvLogger(filePath string) (*CsvLogger, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	return &CsvLogger{
		file:   file,
		writer: writer,
	}, nil
}

// Log writes a log entry to the CSV file and flushes it immediately
func (logger *CsvLogger) Log(record []string) error {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	if err := logger.writer.Write(record); err != nil {
		return err
	}
	logger.writer.Flush()
	return logger.writer.Error()
}

func (logger *CsvLogger) MustLog(record []string) {
	if err := logger.Log(record); err != nil {
		panic(err)
	}
}

// Close cleans up resources used by CsvLogger
func (logger *CsvLogger) Close() error {
	logger.writer.Flush()
	if closeErr := logger.file.Close(); closeErr != nil {
		return closeErr
	}
	return logger.writer.Error()
}
