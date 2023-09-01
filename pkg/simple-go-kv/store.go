package simplegokv

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"
	"time"
)

// entry is the object stored insided the map.
// we attach also an expire time field so we can clean up the expired objects
type entry struct {
	ExpireTime time.Time
	Data       []byte
}

type Database map[string]*entry

type kvStore struct {
	ctx       context.Context
	mutex     *sync.RWMutex
	dataStore Database
}

func NewKVStore() SimpleKV {
	return &kvStore{
		ctx:       context.Background(),
		mutex:     &sync.RWMutex{},
		dataStore: make(Database),
	}
}

func (k kvStore) Get(key string) ([]byte, bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	e, ok := k.dataStore[key]
	//TODO: add a check if the entry is expired
	//if the entry is expired, trigger a background clean
	if ok {
		return e.Data, true
	}
	return nil, false
}

func (k kvStore) Has(key string) bool {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	_, ok := k.dataStore[key]
	return ok
}

func (k kvStore) Set(key string, value any, ttl *int) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	bts, err := k.serialize(value)
	if err != nil {
		return err
	}
	e := entry{Data: bts}
	if ttl != nil {
		exp := time.Now().Add(time.Millisecond * time.Duration(*ttl))
		e.ExpireTime = exp
	}
	k.dataStore[key] = &e
	return nil
}

func (k kvStore) Delete(key string) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	delete(k.dataStore, key)

}

// should pass the input, and the output should be a pointer to a object
func (k kvStore) Deserialize(input []byte, output interface{}) error {
	buf := &bytes.Buffer{}
	buf.Write(input)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(output)
	if err != nil {
		return err
	}
	return nil
}
func (k kvStore) serialize(value any) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
