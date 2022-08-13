package logging

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

const (
	logBufferSize = 2_000_000

	LevelFatal logLevel = "FATAL"
	LevelError logLevel = "ERROR"
	LevelWarn  logLevel = "WARN"
	LevelInfo  logLevel = "INFO"
)

const (
	LogTypeError = iota
	LogTypeRequest
)

type (
	Logger struct {
		timeFormat string

		buffer     [logBufferSize]byte
		actualSize int
		bufMx      *sync.RWMutex

		filepath  string
		fileFlags int
		filePerms os.FileMode
		fileMx    *sync.Mutex
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

func NewLogger(filepath string, fileFlags int, filePerms os.FileMode, timeFormat string) Logger {
	return Logger{
		timeFormat: timeFormat,
		buffer:     [logBufferSize]byte{},
		actualSize: 0,
		bufMx:      &sync.RWMutex{},
		filepath:   filepath,
		fileFlags:  fileFlags,
		filePerms:  filePerms,
		fileMx:     &sync.Mutex{},
	}
}

func (l *Logger) write(data []byte) error {
	l.bufMx.Lock()
	defer l.bufMx.Unlock()

	dataLen := len(data)

	if logBufferSize-l.actualSize < dataLen {
		l.bufMx.Unlock()
		err := l.Save()
		l.bufMx.Lock()
		if err != nil {
			return err
		}
		if logBufferSize-l.actualSize < dataLen {
			return errNoSpace
		}
	}

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

func (l *Logger) Save() error {
	l.fileMx.Lock()
	defer l.fileMx.Unlock()

	f, err := os.OpenFile(l.filepath, l.fileFlags, l.filePerms)
	if err != nil {
		return err
	}
	err = l.move(f)
	_ = f.Close()
	return err
}

func (l *Logger) WriteServerRequest(req Request, res Response) {
	//requestBodyHash := md5.Sum([]byte(req.Body))
	//req.Body = hex.EncodeToString(requestBodyHash[:])

	//responseBodyHash := md5.Sum([]byte(res.Body))
	//res.Body = hex.EncodeToString(responseBodyHash[:])

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

func (l *Logger) WriteError(err error, level logLevel) {
	if err == nil {
		return
	}

	l.writeMessage(err.Error(), level)
}

func (l *Logger) writeMessage(message string, level logLevel) {
	err := l.encodeAndWrite(&ErrorsRecord{
		BaseRecord: BaseRecord{
			Timestamp: time.Now().Format(l.timeFormat),
			Level:     string(level),
			Type:      LogTypeError,
		},
		Content: message,
	})

	if err != nil {
		log.Printf("ERROR LOGGING REQUEST: %v\n", err)
	}
}
