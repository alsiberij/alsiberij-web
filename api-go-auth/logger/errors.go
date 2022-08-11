package logger

import (
	"encoding/json"
	"time"
)

type (
	ErrorsRecord struct {
		BaseRecord
		Content string `json:"content"`
	}
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

	content, _ := json.Marshal(&ErrorsRecord{
		BaseRecord: BaseRecord{
			Timestamp: time.Now().Format(TimeFormat),
			Level:     string(lvl),
			Type:      LogTypeError,
		},
		Content: message,
	})

	writeLog(content)
}
