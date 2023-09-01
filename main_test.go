package main_test

import (
	"fmt"
	"testing"

	simplegokv "github.com/s0und0fs1lence/simple-go-kv/pkg/simple-go-kv"
)

func BenchmarkSet(b *testing.B) {
	kv := simplegokv.NewKVStore(8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := []byte(fmt.Sprintf("value%d", i))
		kv.Set(key, value, nil)
	}
}
func BenchmarkGet(b *testing.B) {
	kv := simplegokv.NewKVStore(8)
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		value := []byte(fmt.Sprintf("value%d", i))
		ttl := 1000000
		kv.Set(key, value, &ttl)
		keys[i] = key
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		kv.Get(keys[i])
	}
}
