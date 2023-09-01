package main

import (
	"fmt"
	"log"

	simplegokv "github.com/s0und0fs1lence/simple-go-kv/pkg/simple-go-kv"
)

func main() {
	kv := simplegokv.NewKVStore(8)
	flpath := "test.rdb"
	kv.Load(&flpath)
	ttl := 10000000 //1 second
	for i := 0; i < ttl; i++ {
		err := kv.Set(fmt.Sprintf("kv_%d", i), fmt.Sprintf("val_%d", i), &ttl)
		if err != nil {
			log.Panic(err)
		}
	}

	err := kv.Save(&flpath)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Done generating file")

}
