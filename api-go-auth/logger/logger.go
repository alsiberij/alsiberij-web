package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const (
	LevelFatal logLevel = "FATAL"
	LevelError logLevel = "ERROR"
	LevelWarn  logLevel = "WARN"
	LevelInfo  logLevel = "INFO"

	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02T15:04:05"

	FilePermissions = 0666

	FilenamePatternLogs = "logs_%s.log"

	IndexPatternLogs = "log-api-go-auth-server-%s"
)

const (
	LogTypeError = iota
	LogTypeRequest
)

type (
	logLevel string

	ElasticSearchConfig struct {
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	}

	BaseRecord struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"logLevel"`
		Type      int    `json:"type"`
	}
)

var (
	LogsPath string

	fileMutex sync.Mutex
)

func writeLog(content []byte) {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	f, err := os.OpenFile(fmt.Sprintf(LogsPath+"/"+FilenamePatternLogs, time.Now().Format(DateFormat)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, FilePermissions)
	if err != nil {
		log.Printf("FAILED TO WRITE LOG: %s", err.Error())
		return
	}

	_, err = f.Write(content)
	_, err = f.Write([]byte{'\n'})
	if err != nil {
		log.Printf("FAILED TO WRITE LOG: %s", err.Error())
	}

	_ = f.Close()
}
