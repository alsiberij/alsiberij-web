package utils

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

var (
	R = syncedRandom{
		rand: rand.New(rand.NewSource(time.Now().Unix())),
		mx:   sync.Mutex{},
	}
)

type (
	syncedRandom struct {
		rand *rand.Rand
		mx   sync.Mutex
	}

	Searchable interface {
		string
	}
)

func (s *syncedRandom) IntSync() int {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.rand.Int()
}

func (s *syncedRandom) Int() int {
	return s.rand.Int()
}

func (s *syncedRandom) Acquire() {
	s.mx.Lock()
}

func (s *syncedRandom) Release() {
	s.mx.Unlock()
}

func GenerateString(length uint, alphabet string) string {
	result := make([]byte, length)

	R.Acquire()
	defer R.Release()

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
