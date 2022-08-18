package utils

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

var (
	R = NewSyncedRandom(time.Now().Unix())
)

type (
	SyncedRandom struct {
		rand *rand.Rand
		mx   sync.Mutex
	}

	Searchable interface {
		string
	}
)

func NewSyncedRandom(seed int64) SyncedRandom {
	return SyncedRandom{
		rand: rand.New(rand.NewSource(seed)),
		mx:   sync.Mutex{},
	}
}

func (s *SyncedRandom) Int() int {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.rand.Int()
}

func (s *SyncedRandom) int() int {
	return s.rand.Int()
}

func (s *SyncedRandom) acquire() {
	s.mx.Lock()
}

func (s *SyncedRandom) release() {
	s.mx.Unlock()
}

func GenerateString(length uint, alphabet string) string {
	result := make([]byte, length)

	R.acquire()
	defer R.release()

	for i := range result {
		result[i] = alphabet[R.int()%len(alphabet)]
	}
	return BytesToString(result)
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
	return *(*string)(unsafe.Pointer(&slice))
}

func GenerateCode(length uint) string {
	result := make([]byte, length)

	R.acquire()
	defer R.release()

	for i := range result {
		result[i] = byte(R.int()%10 + 48)
	}
	return BytesToString(result)
}
