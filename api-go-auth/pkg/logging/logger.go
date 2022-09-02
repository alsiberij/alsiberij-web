package logging

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
)

const (
	LogTypeError = iota
	LogTypeRequest
)

//TODO Ticker, simplify writing

type (
	Logger struct {
		timeFormat string

		buffer     []byte
		actualSize int
		maxSize    int
		bufMx      *sync.RWMutex

		filepath  string
		fileFlags int
		filePerms os.FileMode
		fileMx    *sync.Mutex

		currentDate string
	}

	logLevel string
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

var (
	errNoSpace = errors.New("not enough space to write")
)

func NewLogger(bufferSize int, filepath string, fileFlags int, filePerms os.FileMode, timeFormat string) Logger {
	return Logger{
		timeFormat:  timeFormat,
		buffer:      make([]byte, bufferSize),
		actualSize:  0,
		maxSize:     bufferSize,
		bufMx:       &sync.RWMutex{},
		filepath:    filepath,
		fileFlags:   fileFlags,
		filePerms:   filePerms,
		fileMx:      &sync.Mutex{},
		currentDate: time.Now().Format("2006-01-02"),
	}
}

func (l *Logger) write(data []byte) error {
	dataLen := len(data)
	actualDate := time.Now().Format("2006-01-02")

	if actualDate != l.currentDate || l.maxSize-l.actualSize < dataLen+1 {
		err := l.save(l.currentDate)
		if err != nil {
			return err
		}
		if l.maxSize-l.actualSize < dataLen {
			return errNoSpace
		}
		l.currentDate = actualDate
	}

	l.bufMx.Lock()
	defer l.bufMx.Unlock()

	for i := 0; i < dataLen; i++ {
		l.buffer[l.actualSize+i] = data[i]
	}
	l.buffer[l.actualSize+dataLen] = '\n'
	l.actualSize += dataLen + 1

	return nil
}

func (l *Logger) move(w io.Writer) error {
	l.bufMx.RLock()
	defer l.bufMx.RUnlock()

	if l.actualSize == 0 {
		return nil
	}
	_, err := w.Write(l.buffer[:l.actualSize])
	l.actualSize = 0
	return err
}

func (l *Logger) encodeAndWrite(data interface{}) error {
	content, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return l.write(content)
}

func (l *Logger) save(date string) error {
	l.fileMx.Lock()
	defer l.fileMx.Unlock()

	f, err := os.OpenFile(fmt.Sprintf(l.filepath, date), l.fileFlags, l.filePerms)
	if err != nil {
		return err
	}
	err = l.move(f)
	_ = f.Close()
	return err
}

func (l *Logger) Save() error {
	return l.save(time.Now().Format("2006-01-02"))
}

func (l *Logger) WriteServerRequest(req Request, res Response) {
	//TODO Hash bodies

	//requestBodyHash := md5.Sum([]byte(req.Body))
	//req.Body = hex.EncodeToString(requestBodyHash[:])

	//responseBodyHash := md5.Sum([]byte(res.Body))
	//res.Body = hex.EncodeToString(responseBodyHash[:])

	req.Body = base64.URLEncoding.EncodeToString([]byte(req.Body))
	res.Body = base64.URLEncoding.EncodeToString([]byte(res.Body))

	err := l.encodeAndWrite(&ServerRecord{
		BaseRecord: BaseRecord{
			Timestamp: time.Now().Format(l.timeFormat),
			Level:     string(LevelInfo),
			Type:      LogTypeRequest,
		},
		Content: serverRecordContent{
			Request:  &req,
			Response: &res,
		},
	})

	if err != nil {
		log.Printf("ERROR LOGGING REQUEST: %v\n", err)
	}
}

func (l *Logger) LogError(err error, level logLevel) {
	if err == nil {
		return
	}

	err = l.encodeAndWrite(&ErrorsRecord{
		BaseRecord: BaseRecord{
			Timestamp: time.Now().Format(l.timeFormat),
			Level:     string(level),
			Type:      LogTypeError,
		},
		Content: err.Error(),
	})

	if err != nil {
		log.Printf("ERROR LOGGING REQUEST: %v\n", err)
	}
}
