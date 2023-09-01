package main

import (
	"log"

	simplegokv "github.com/s0und0fs1lence/simple-go-kv/pkg/simple-go-kv"
)

func main() {
	kv := simplegokv.NewKVStore()
	key := "TEST_1"
	expected := map[string]interface{}{"k": 1}
	set_error := kv.Set(key, expected, nil)
	if set_error != nil {
		log.Panic(set_error)
	}
	bts, success := kv.Get(key)
	if !success {
		log.Panic("something is weird!")
	}
	var comp map[string]interface{}
	err := kv.Deserialize(bts, &comp)
	if err != nil {
		log.Panic(err)
	}
	if comp["k"] != expected["k"] {
		log.Panic("DIFFERENT RESULT")
	}
	log.Printf("ALL TEST PASSED")

}
