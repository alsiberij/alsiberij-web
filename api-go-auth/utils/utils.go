package utils

import (
	"math/rand"
	"sync"
	"unsafe"
)

type (
	Random struct {
		rand *rand.Rand
		mx   *sync.Mutex
	}

	Searchable interface {
		string
	}
)

func NewRandom(seed int64) Random {
	return Random{
		rand: rand.New(rand.NewSource(seed)),
		mx:   &sync.Mutex{},
	}
}

func (r *Random) Int() int {
	if r.mx == nil || r.rand == nil {
		return 0
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	return r.rand.Int()
}

func (r *Random) String(length uint, alphabet string) string {
	if r.mx == nil || r.rand == nil {
		return ""
	}

	result := make([]byte, length)

	r.mx.Lock()
	defer r.mx.Unlock()

	for i := range result {
		result[i] = alphabet[r.rand.Int()%len(alphabet)]
	}
	return BytesToString(result)
}

func (r *Random) Code(length uint) string {
	if r.mx == nil || r.rand == nil {
		return ""
	}

	result := make([]byte, length)

	r.mx.Lock()
	defer r.mx.Unlock()

	for i := range result {
		result[i] = byte(r.rand.Int()%10 + '0')
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
