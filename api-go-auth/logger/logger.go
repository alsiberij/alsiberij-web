package logger

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
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

	elasticsearchClient *elasticsearch.Client

	fileMutex sync.Mutex
)

func Init(config ElasticSearchConfig) error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			fmt.Sprintf("%s://%s:%d", config.Protocol, config.Host, config.Port),
		},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	_, err = client.Ping()
	if err != nil {
		return err
	}

	elasticsearchClient = client

	return nil
}

func writeLog(content []byte) {
	rq := esapi.IndexRequest{Index: fmt.Sprintf(IndexPatternLogs, time.Now().Format(DateFormat)),
		Body: bytes.NewReader(content)}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := rq.Do(ctx, elasticsearchClient)
	if err != nil {
		log.Printf("FAILED TO WRITE LOG: %s", err.Error())
	} else {
		if r.IsError() {
			log.Printf("FAILED TO WRITE LOG: %s", r.String())
		}
		_ = r.Body.Close()
	}

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
