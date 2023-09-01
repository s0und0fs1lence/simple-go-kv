package simplegokv

import "sync"

// shard is a shard of the KV store.
type shard struct {
	dataStore map[string]*entry
	mutex     sync.RWMutex
}

func fnv1aHash(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= 16777619
	}
	return hash
}

// getShard returns the shard responsible for a given key.
func (k *kvStore) getShard(key string) *shard {
	hash := fnv1aHash(key) // Use a hash function to determine the shard
	return k.shards[hash%uint32(len(k.shards))]
}
