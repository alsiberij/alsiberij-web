package utils

import (
	"math/rand"
	"time"
	"unsafe"
)

var (
	R = rand.New(rand.NewSource(time.Now().Unix()))
)

func GenerateString(length uint, alphabet string) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = alphabet[R.Int()%len(alphabet)]
	}
	return *(*string)(unsafe.Pointer(&result))
}
