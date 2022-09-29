package logging

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	DefaultBufferSize      = 1_000_000
	DefaultFilenamePattern = "logs-%s.log"
	DefaultFlags           = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	DefaultPerms           = 0777
	DefaultTimeFormat      = "2006-01-02T15:04:05"
	DefaultSaveInterval    = time.Second * 3

	LevelFatal logLevel = "FATAL"
	LevelError logLevel = "ERROR"
	LevelInfo  logLevel = "INFO"
)

const (
	LogTypeError = iota
	LogTypeRequest
)

type (
	BaseRecord struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"logLevel"`
		Type      int    `json:"type"`
	}
	ServerRecord struct {
		BaseRecord
		Content serverRecordContent `json:"content"`
	}
	ErrorsRecord struct {
		BaseRecord
		Content string `json:"content"`
	}

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
		Request  *Request  `json:"request"`
		Response *Response `json:"response"`
	}
)

type (
	Logger struct {
		timeFormat  string
		currentDate string

		saveInterval time.Duration

		buffer     []byte
		actualSize int
		maxSize    int

		filenamePattern string
		fileFlags       int
		filePerms       os.FileMode

		inputCh chan []byte
		errCh   chan error

		wg *sync.WaitGroup
	}

	logLevel string
)

var (
	ErrNoSpace = errors.New("not enough space to write")
	ErrClosed  = errors.New("logger closed")
)

func NewLogger(bufferSize int, filepath string, fileFlags int, filePerms os.FileMode, timeFormat string, saveInterval time.Duration) *Logger {
	l := &Logger{
		timeFormat:      timeFormat,
		saveInterval:    saveInterval,
		buffer:          make([]byte, bufferSize),
		actualSize:      0,
		maxSize:         bufferSize,
		filenamePattern: filepath,
		fileFlags:       fileFlags,
		filePerms:       filePerms,
		currentDate:     time.Now().Format("2006-01-02"),
		inputCh:         make(chan []byte),
		errCh:           make(chan error),
		wg:              &sync.WaitGroup{},
	}

	go l.logWorker()

	return l
}

func (l *Logger) write(data []byte) error {
	l.wg.Add(1)
	defer l.wg.Done()

	dataLen := len(data)
	actualDate := time.Now().Format("2006-01-02")

	if actualDate != l.currentDate || l.maxSize-l.actualSize < dataLen+1 {
		err := l.save(l.currentDate)
		if err != nil {
			return err
		}
		if l.maxSize-l.actualSize < dataLen {
			return ErrNoSpace
		}
		l.currentDate = actualDate
	}

	for i := 0; i < dataLen; i++ {
		l.buffer[l.actualSize+i] = data[i]
	}
	l.buffer[l.actualSize+dataLen] = '\n'
	l.actualSize += dataLen + 1

	return nil
}

func (l *Logger) save(date string) error {
	l.wg.Add(1)
	defer l.wg.Done()

	if l.actualSize == 0 {
		return nil
	}

	f, err := os.OpenFile(fmt.Sprintf(l.filenamePattern, date), l.fileFlags, l.filePerms)
	if err != nil {
		return err
	}
	_, err = f.Write(l.buffer[:l.actualSize])
	if err == nil {
		l.actualSize = 0
	}
	_ = f.Close()
	return err
}

func (l *Logger) WriteServerRequest(req Request, res Response) error {
	//TODO Hash bodies

	//requestBodyHash := md5.Sum([]byte(req.Body))
	//req.Body = hex.EncodeToString(requestBodyHash[:])

	//responseBodyHash := md5.Sum([]byte(res.Body))
	//res.Body = hex.EncodeToString(responseBodyHash[:])

	req.Body = base64.URLEncoding.EncodeToString([]byte(req.Body))
	res.Body = base64.URLEncoding.EncodeToString([]byte(res.Body))

	record := &ServerRecord{
		BaseRecord: BaseRecord{
			Timestamp: time.Now().Format(l.timeFormat),
			Level:     string(LevelInfo),
			Type:      LogTypeRequest,
		},
		Content: serverRecordContent{
			Request:  &req,
			Response: &res,
		},
	}
	content, _ := json.Marshal(record)

	l.inputCh <- content
	return <-l.errCh
}

func (l *Logger) WriteError(err error, level logLevel) error {
	record := &ErrorsRecord{
		BaseRecord: BaseRecord{
			Timestamp: time.Now().Format(l.timeFormat),
			Level:     string(level),
			Type:      LogTypeError,
		},
		Content: err.Error(),
	}
	content, _ := json.Marshal(record)

	l.inputCh <- content

	err, ok := <-l.errCh
	if !ok {
		err = ErrClosed
	}

	return err
}

func (l *Logger) Close() error {
	l.wg.Wait()
	close(l.inputCh)
	return l.save(time.Now().Format("2006-01-02"))
}

func (l *Logger) logWorker() {
	defer close(l.errCh)

	for {
		select {
		case data, ok := <-l.inputCh:
			if !ok {
				return
			}
			l.errCh <- l.write(data)
		case <-time.After(l.saveInterval):
			err := l.save(time.Now().Format("2006-01-02"))
			if err != nil {
				return
			}
		}
	}
}
