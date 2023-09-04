package simplegokv

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// serializedEntry represents a serialized key-value entry.
type serializedEntry struct {
	Key        string
	ExpireTime time.Time
	Data       []byte
}

// deserializeEntry deserializes a SerializedEntry from a buffer.
func deserializeEntry(data []byte) (serializedEntry, error) {
	if len(data) < 4 {
		return serializedEntry{}, fmt.Errorf("invalid entry data")
	}

	keyLength := binary.BigEndian.Uint32(data[0:4])
	if len(data) < int(keyLength+12) {
		return serializedEntry{}, fmt.Errorf("invalid entry data")
	}

	key := string(data[4 : 4+keyLength])
	expireTimeUnix := int64(binary.BigEndian.Uint64(data[4+keyLength : 12+keyLength]))
	expireTime := time.Unix(expireTimeUnix, 0)
	dataLength := binary.BigEndian.Uint32(data[12+keyLength : 16+keyLength])
	entryData := data[16+keyLength : 16+keyLength+uint32(dataLength)]

	return serializedEntry{
		Key:        key,
		ExpireTime: expireTime,
		Data:       entryData,
	}, nil
}

// TODO: add single read for entire entry
// TODO: add decompression logic
func (k kvStore) Load(filename *string) error {
	return fmt.Errorf("Load function is currently not implemented")
	var flName string
	if filename != nil {
		flName = *filename
	} else {
		flName = k.baseFileName
	}
	file, err := os.Open(flName)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	var version uint64
	if err := binary.Read(gzipReader, binary.BigEndian, &version); err != nil {
		return err
	}

	if version != 1 {
		return fmt.Errorf("unsupported RDB version: %d", version)
	}
	errorCount := 0
	// Read keys and values from the RDB file and populate the KV store
	for {
		// Read entry length
		var entryLength uint32
		if err := binary.Read(gzipReader, binary.BigEndian, &entryLength); err != nil {
			if err == io.EOF {
				break // End of file
			}
			return err
		}

		// Read the entire entry into a buffer
		entryBuffer := make([]byte, entryLength)
		if _, err := io.ReadFull(gzipReader, entryBuffer); err != nil {
			if err == io.EOF {
				break // End of file
			}
			return err
		}

		// Deserialize the entry from the buffer
		serializedEntry, err := deserializeEntry(entryBuffer)
		if err != nil {
			log.Printf("got error while deserializing data  %s", err)
			errorCount += 1
			// Handle error
			continue
		}

		// Insert the key-value pair into the KV store
		err = k.setOnLoad(serializedEntry.Key, serializedEntry.Data, serializedEntry.ExpireTime)
		if err != nil {
			log.Printf("got error while inserting key %s: %s", serializedEntry.Key, err)
			errorCount += 1
		}
	}
	log.Printf("[LOAD-RDB] Successfull loaded file, with %d errors", errorCount)

	return nil
}

// TODO: try to speed up the process
// TODO: define a serialization format in some sort of interface, so the deserializer, can use the interface to retrieve the informations
func (k kvStore) Save(filename *string) error {
	return fmt.Errorf("Save function is currently not implemented")

	var flName string
	if filename != nil {
		flName = *filename
	} else {
		flName = k.baseFileName
	}
	file, err := os.Create(flName)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip writer on top of a buffered writer for efficient writes
	bufferedWriter := bufio.NewWriter(file)
	gzipWriter := gzip.NewWriter(bufferedWriter)
	defer gzipWriter.Close()

	// Write the RDB version (a simple version 1 for this example)
	if err := binary.Write(gzipWriter, binary.BigEndian, uint64(1)); err != nil {
		return err
	}

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Define a channel to send serialized entries
	entriesCh := make(chan serializedEntry, len(k.shards))

	// Iterate over all shards and process them concurrently
	for _, sh := range k.shards {
		wg.Add(1)
		go func(sh *shard) {
			defer wg.Done()

			sh.mutex.RLock()
			defer sh.mutex.RUnlock()

			// Iterate over each entry in the shard
			for key, entry := range sh.dataStore {
				if entry.isNotExpired() {
					// Serialize the entry and send it to the channel
					serializedEntry := serializedEntry{
						Key:        key,
						ExpireTime: entry.ExpireTime,
						Data:       entry.Data,
					}
					entriesCh <- serializedEntry
				}
			}
		}(sh)
	}

	// Close the channel after all goroutines finish
	go func() {
		wg.Wait()
		close(entriesCh)
	}()

	// Process serialized entries from the channel and write them to the gzip writer
	for serializedEntry := range entriesCh {
		// Serialize key length, key, expiration time, and data length
		keyLength := uint32(len(serializedEntry.Key))
		dataLength := uint32(len(serializedEntry.Data))
		entryBuffer := make([]byte, 4+len(serializedEntry.Key)+8+4)

		binary.BigEndian.PutUint32(entryBuffer[0:4], keyLength)
		copy(entryBuffer[4:], []byte(serializedEntry.Key))
		binary.BigEndian.PutUint64(entryBuffer[4+keyLength:12+keyLength], uint64(serializedEntry.ExpireTime.Unix()))
		binary.BigEndian.PutUint32(entryBuffer[12+keyLength:16+keyLength], dataLength)

		// Write the serialized entry to the gzip writer
		if _, err := gzipWriter.Write(entryBuffer); err != nil {
			return err
		}
		if _, err := gzipWriter.Write(serializedEntry.Data); err != nil {
			return err
		}
	}

	// Flush the buffered writer to ensure all data is written to the file
	if err := bufferedWriter.Flush(); err != nil {
		return err
	}

	return nil
}
