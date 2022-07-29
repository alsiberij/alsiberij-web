package utils

import (
	"math/rand"
	"time"
	"unsafe"
)

var (
	R = rand.New(rand.NewSource(time.Now().Unix()))
)

type (
	Searchable interface {
		string
	}
)

func GenerateString(length uint, alphabet string) string {
	result := make([]byte, length)
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
