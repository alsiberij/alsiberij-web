package logger

import (
	"encoding/json"
	"log"
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
	if elasticsearchClient.Conn == nil {
		log.Println("ELASTICSEARCH CLIENT IS NOT ALIVE")
		return
	}

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
