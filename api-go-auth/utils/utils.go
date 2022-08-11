package utils

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

//TODO wrap r with mutex

var (
	R      = rand.New(rand.NewSource(time.Now().Unix()))
	rMutex sync.Mutex
)

type (
	Searchable interface {
		string
	}
)

func GenerateString(length uint, alphabet string) string {
	result := make([]byte, length)
	rMutex.Lock()
	defer rMutex.Unlock()
	for i := range result {
		result[i] = alphabet[R.Int()%len(alphabet)]
	}
	return *(*string)(unsafe.Pointer(&result))
}

func ExistsIn[T Searchable](haystack []T, needle T) bool {
	for i := range haystack {
		if haystack[i] == needle {
			return true
		}
	}
	return false
}

func BytesToString(slice []byte) string {
	return *((*string)(unsafe.Pointer(&slice)))
}
