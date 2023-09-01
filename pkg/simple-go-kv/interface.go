package simplegokv

// SimpleKV is the interface we should implement to have a "functional" key value store
type SimpleKV interface {
	Get(key string) ([]byte, bool)
	Has(key string) bool
	Set(key string, value any, ttl *int) error
	Delete(key string)
	Deserialize(input []byte, output interface{}) error
	Load(filename *string) error
	Save(filename *string) error
	TruncateDatabase()
}
