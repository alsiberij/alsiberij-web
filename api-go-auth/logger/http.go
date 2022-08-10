package logger

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	FilenamePatternForRequests = "requests_%s.log"
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
	ServerRecord struct {
		Timestamp int64               `json:"timestamp"`
		Level     string              `json:"level"`
		Type      int                 `json:"type"`
		Content   serverRecordContent `json:"content"`
	}
	serverRecordContent struct {
		Request  Request  `json:"request"`
		Response Response `json:"response"`
	}
)

func LogServerRequest(req Request, res Response) {
	requestBodyHash := md5.Sum([]byte(req.Body))
	req.Body = hex.EncodeToString(requestBodyHash[:])

	responseBodyHash := md5.Sum([]byte(res.Body))
	res.Body = hex.EncodeToString(responseBodyHash[:])

	content, _ := json.Marshal(&ServerRecord{
		Timestamp: time.Now().Unix(),
		Level:     string(LevelInfo),
		Type:      LogTypeRequest,
		Content: serverRecordContent{
			Request:  req,
			Response: res,
		},
	})

	f, err := os.OpenFile(fmt.Sprintf(LogsPath+"/"+FilenamePatternForRequests, time.Now().Format(FilenameDateFormat)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0222)
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
