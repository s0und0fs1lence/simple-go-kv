# simple-go-kv - A Lightweight KeyValue Store in Go

simple-go-kv is a simple and lightweight key-value store library written in Go. It provides an easy-to-use interface for storing and retrieving data using keys. With optional time-to-live (TTL) support, you can set an expiration time for your data, making it ideal for use cases like caching.

## Features

- **Easy-to-Use**: SimpleGoKV offers a straightforward API for storing, retrieving, and deleting key-value pairs.
- **Optional TTL**: Set expiration times for your data, allowing you to automatically clean up expired entries.
- **Concurrency-Safe**: The library uses mutexes to ensure safe concurrent access to your data.

## Roadmap

Here's a list of features and improvements planned for SimpleGoKV. Feel free to contribute or check the progress on these items:

- [ ] Implement additional caching strategies.
- [ ] Add support for data expiration notifications.
- [ ] Improve error handling and error reporting.
- [ ] Provide more examples and documentation.
- [ ] Optimize performance for high-throughput scenarios.
- [ ] Add unit tests for critical components.
- [ ] Add support for shards
- [ ] Add persistence of the data across reboots

### Completed Features

- [x] Basic key-value storage functionality.
- [x] Optional time-to-live (TTL) support.
- [x] Concurrency-safe data access.

If you have ideas for new features or improvements, please open an issue to discuss them or consider contributing by submitting a pull request.


## Installation

To use SimpleGoKV in your Go project, you can install it using `go get`:

```shell
go get github.com/s0und0fs1lence/simple-go-kv
```


## Usage

Here's a quick example of how to use SimpleGoKV:

```go
package main

import (
	"fmt"
	simplegokv "github.com/s0und0fs1lence/simple-go-kv"
)

func main() {
	// Create a new KeyValue store
	store := simplegokv.NewKVStore()

	// Set a key-value pair
	store.Set("name", "John", nil)

	// Get a value by key
	value, found := store.Get("name")
	if found {
		fmt.Println("Name:", string(value))
	} else {
		fmt.Println("Name not found.")
	}
}
```


## API

- `NewKVStore() SimpleKV`: Create a new instance of the KeyValue store.
- `Get(key string) ([]byte, bool)`: Retrieve a value by key. Returns the value and a boolean indicating if the key was found.
- `Has(key string) bool`: Check if a key exists in the store.
- `Set(key string, value any, ttl *int) error`: Set a key-value pair. You can provide an optional TTL (time-to-live) in milliseconds.
- `Delete(key string)`: Delete a key-value pair.
- `Deserialize(input []byte, output interface{}) error`: Deserialize a byte slice into an object.

## Contributing

Contributions to simple-go-kv are welcome! If you have any feature requests, bug reports, or improvements, please open an issue or submit a pull request on the [GitHub repository](https://github.com/s0und0fs1lence/simple-go-kv). 

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

Created by [s0und0fs1lence](https://github.com/s0und0fs1lence). You can find me on GitHub, and I'm open to collaboration and feedback. Feel free to reach out!
