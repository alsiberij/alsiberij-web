package logging

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const concurrency = 10000

type (
	logEntry struct {
		Type int `json:"type"`
	}
	exampleBody struct {
		Id int `json:"id"`
	}
)

func closeAndDeleteFile(f *os.File) {
	if f != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}
}

func TestLogger(t *testing.T) {
	l := NewLogger(DefaultBufferSize, DefaultFilenamePattern, DefaultFlags, DefaultPerms, DefaultTimeFormat, DefaultSaveInterval)
	var wg sync.WaitGroup

	err := l.WriteError(errors.New("#0"), LevelInfo)
	if err != nil {
		t.Fatalf("UNEXPECTED LOG ERROR: %v", err)
	}

	time.Sleep(DefaultSaveInterval + time.Second*3)

	filename := fmt.Sprintf(DefaultFilenamePattern, time.Now().Format("2006-01-02"))
	_, err = os.Stat(filename)
	if err != nil {
		t.Fatal("LOG AUTO SAVE DOES NOT WORK")
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		if i%2 == 0 {
			go func(l *Logger, i int, wg *sync.WaitGroup) {
				defer wg.Done()

				err = l.WriteError(errors.New(fmt.Sprintf("%d", i)), LevelInfo)
				if err != nil {
					t.Errorf("UNEXPECTED ERROR FROM #%d: %v", i, err)
					return
				}
			}(l, i, &wg)
		} else {
			go func(l *Logger, i int, wg *sync.WaitGroup) {
				defer wg.Done()

				body, _ := json.Marshal(exampleBody{Id: i})
				ts := time.Now().Unix()

				err = l.WriteServerRequest(Request{
					Timestamp: ts,
					Method:    "TEST",
					Path:      "/test",
					Protocol:  "TP",
					Headers:   []string{"Test: test"},
					Body:      string(body),
				}, Response{
					Timestamp:     ts + 1,
					Protocol:      "TEST",
					StatusCode:    200,
					Headers:       []string{"Status: OK"},
					Body:          `{"status":"OK"}`,
					ExecutionTime: 1,
				})
				if err != nil {
					t.Errorf("UNEXPECTED ERROR FROM #%d: %v", i, err)
					return
				}
			}(l, i, &wg)
		}
	}

	wg.Wait()

	err = l.Close()
	if err != nil {
		t.Fatalf("UNEXPECTED ERROR SAVE: %v", err)
	}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0777)
	if err != nil {
		t.Fatalf("UNEXPECTED LOG OPEN ERROR: %v", err)
	}
	defer closeAndDeleteFile(f)

	goroutineStatuses := make([]bool, concurrency)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		content := scanner.Bytes()

		var row logEntry
		err = json.Unmarshal(content, &row)
		if err != nil {
			t.Fatalf("UNEXPECTED UNMARSHAL ERROR: %v", err)
		}

		var goroutineId int
		if row.Type == LogTypeError {
			var record ErrorsRecord
			err = json.Unmarshal(content, &record)
			if err != nil {
				t.Fatalf("UNEXPECTED UNMARSHAL ERROR: %v", err)
			}
			goroutineId, _ = strconv.Atoi(record.Content)
		} else if row.Type == LogTypeRequest {
			var record ServerRecord
			err = json.Unmarshal(content, &record)
			if err != nil {
				t.Fatalf("UNEXPECTED UNMARSHAL ERROR: %v", err)
			}
			var body exampleBody
			err = json.Unmarshal([]byte(record.Content.Request.Body), &body)
			if err != nil {
				t.Fatalf("UNEXPECTED UNMARSHAL ERROR: %v", err)
			}
			goroutineId = body.Id
		}
		goroutineStatuses[goroutineId] = true
	}

	for i := range goroutineStatuses {
		if !goroutineStatuses[i] {
			t.Fatalf("LOG ENTRY #%d NOT FOUND", i)
		}
	}
}
