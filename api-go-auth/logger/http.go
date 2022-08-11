package logger

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"
)

type (
	Request struct {
		Timestamp int64    `json:"timestamp"`
		Method    string   `json:"method"`
		Path      string   `json:"path"`
		Protocol  string   `json:"protocol"`
		Headers   []string `json:"headers"`
		Body      string   `json:"body"`
	}
	Response struct {
		Timestamp     int64    `json:"timestamp"`
		Protocol      string   `json:"protocol"`
		StatusCode    int      `json:"statusCode"`
		Headers       []string `json:"headers"`
		Body          string   `json:"body"`
		ExecutionTime int64    `json:"executionTime"`
	}
	serverRecordContent struct {
		Request  Request  `json:"request"`
		Response Response `json:"response"`
	}

	ServerRecord struct {
		BaseRecord
		Content serverRecordContent `json:"content"`
	}
)

func LogServerRequest(req Request, res Response) {
	if elasticsearchClient.Conn == nil {
		log.Println("ELASTICSEARCH CLIENT IS NOT ALIVE")
		return
	}

	requestBodyHash := md5.Sum([]byte(req.Body))
	req.Body = hex.EncodeToString(requestBodyHash[:])

	responseBodyHash := md5.Sum([]byte(res.Body))
	res.Body = hex.EncodeToString(responseBodyHash[:])

	content, _ := json.Marshal(&ServerRecord{
		BaseRecord: BaseRecord{
			Timestamp: time.Now().Format(TimeFormat),
			Level:     string(LevelInfo),
			Type:      LogTypeRequest,
		},
		Content: serverRecordContent{
			Request:  req,
			Response: res,
		},
	})

	writeLog(content)
}
