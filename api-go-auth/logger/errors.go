package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const (
	FilenamePatternForErrors = "errors_%s.log"
)

type (
	RecordErrors struct {
		Timestamp int64  `json:"timestamp"`
		Level     string `json:"logLevel"`
		Type      int    `json:"type"`
		Content   string `json:"content"`
	}
)

var (
	errorMutex sync.Mutex
)

func LogError(err error, lvl logLevel) {
	if err == nil {
		return
	}

	LogMessage(err.Error(), lvl)
}

func LogMessage(message string, lvl logLevel) {
	if message == "" {
		return
	}

	content, _ := json.Marshal(&RecordErrors{
		Timestamp: time.Now().Unix(),
		Level:     string(lvl),
		Type:      LogTypeError,
		Content:   message,
	})

	errorMutex.Lock()
	defer errorMutex.Unlock()
	f, err := os.OpenFile(fmt.Sprintf(LogsPath+"/"+FilenamePatternForErrors, time.Now().Format(FilenameDateFormat)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0222)
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
