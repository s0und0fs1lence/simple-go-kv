package simplegokv

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

var entryPool = sync.Pool{
	New: func() interface{} {
		return &entry{}
	},
}

// entry is the object stored insided the map.
// we attach also an expire time field so we can clean up the expired objects
type entry struct {
	ExpireTime time.Time
	Data       []byte
}

func (e *entry) isNotExpired() bool {
	return e.ExpireTime.IsZero() || time.Now().Before(e.ExpireTime)
}

// dataBase is the map holding the actual data.
type dataBase map[string]*entry

type kvStore struct {
	ctx          context.Context
	baseFileName string
	shards       []*shard
}

func NewKVStore(numShards int) SimpleKV {
	var shards []*shard
	for i := 0; i < numShards; i++ {
		shards = append(shards, &shard{
			dataStore: make(map[string]*entry),
		})
	}

	return &kvStore{
		ctx:          context.Background(),
		baseFileName: "data.rdb",
		shards:       shards,
	}
}

func (k kvStore) Get(key string) ([]byte, bool) {
	shard := k.getShard(key)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()
	e, ok := shard.dataStore[key]
	if ok && e.isNotExpired() {
		return e.Data, true
	}
	return nil, false
}

func (k kvStore) Has(key string) bool {
	shard := k.getShard(key)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()
	e, ok := shard.dataStore[key]
	return ok && e.isNotExpired()
}

func (k kvStore) Set(key string, value any, ttl *int) error {
	shard := k.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	e := entryPool.Get().(*entry)
	bts, err := k.serialize(value)
	if err != nil {
		return err
	}

	e.Data = bts
	if ttl != nil {
		exp := time.Now().Add(time.Millisecond * time.Duration(*ttl))
		e.ExpireTime = exp
	}
	shard.dataStore[key] = e
	return nil
}
func (k kvStore) setOnLoad(key string, value []byte, ttl time.Time) error {
	shard := k.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	e := entry{
		Data:       value,
		ExpireTime: ttl,
	}

	shard.dataStore[key] = &e
	return nil
}

func (k kvStore) Delete(key string) {
	shard := k.getShard(key)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	delete(shard.dataStore, key)

}

func (k kvStore) TruncateDatabase() {
	var wg sync.WaitGroup
	for _, sh := range k.shards {
		wg.Add(1)
		go func(sh *shard) {
			defer wg.Done()
			sh.mutex.Lock()
			for k := range sh.dataStore {
				delete(sh.dataStore, k)
			}
			sh.mutex.Unlock()
		}(sh)

	}
	wg.Wait()

}

func (k kvStore) GetEntryCount() uint32 {
	var totEntry uint32
	var wg sync.WaitGroup
	for _, sh := range k.shards {
		wg.Add(1)
		go func(sh *shard) {
			defer wg.Done()
			sh.mutex.RLock()

			atomic.AddUint32(&totEntry, uint32(len((sh.dataStore))))

			sh.mutex.RUnlock()
		}(sh)

	}
	return totEntry
}
