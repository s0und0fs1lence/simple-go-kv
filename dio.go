package simplegokv

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"
)

// TODO: add single read for entire entry
// TODO: add decompression logic
func (k kvStore) Load(filename *string) error {
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

	bufferedReader := bufio.NewReader(file)
	gzipReader, err := gzip.NewReader(bufferedReader)
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
		var keyLength uint32
		if err := binary.Read(gzipReader, binary.BigEndian, &keyLength); err != nil {
			errorCount += 1
			break
		}
		keyBytes := make([]byte, keyLength)
		if _, err := gzipReader.Read(keyBytes); err != nil {
			errorCount += 1
			break
		}
		key := string(keyBytes)

		var expireTimeUnix int64
		if err := binary.Read(gzipReader, binary.BigEndian, &expireTimeUnix); err != nil {
			errorCount += 1
			break
		}
		expireTime := time.Unix(expireTimeUnix, 0)

		var dataLength uint32
		if err := binary.Read(gzipReader, binary.BigEndian, &dataLength); err != nil {
			errorCount += 1
			break
		}
		data := make([]byte, dataLength)
		if _, err := gzipReader.Read(data); err != nil {
			errorCount += 1
			break
		}

		// Insert the key-value pair into the KV store
		err := k.setOnLoad(key, data, expireTime)
		if err != nil {
			log.Printf("got error while inserting key %s: %s", key, err)
			errorCount += 1
		}
	}
	log.Printf("[LOAD-RDB] Successfull loaded file, with %d errors", errorCount)

	return nil
}

// TODO: try to speed up the process
// TODO: define a serialization format in some sort of interface, so the deserializer, can use the interface to retrieve the informations
func (k kvStore) Save(filename *string) error {
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

	bufferedWriter := bufio.NewWriter(file)
	gzipWriter := gzip.NewWriter(bufferedWriter)
	defer gzipWriter.Close()
	// Write the RDB version (a simple version 1 for this example)
	if err := binary.Write(gzipWriter, binary.BigEndian, uint64(1)); err != nil {
		return err
	}

	// Iterate over all shards and entries and write them to the RDB file
	for _, sh := range k.shards {

		sh.mutex.RLock()
		for key, entry := range sh.dataStore {
			if entry.isNotExpired() {

				// creating a buffer where to store the serialized entry, then write it all at once
				entryBuffer := make([]byte, 0)

				// Serialize the key
				keyLength := uint32(len(key))
				entryBuffer = append(entryBuffer, uint8(keyLength))
				entryBuffer = append(entryBuffer, []byte(key)...)

				// Serialize the expiration time (Unix timestamp in seconds)
				expireTime := entry.ExpireTime.Unix()
				entryBuffer = append(entryBuffer, make([]byte, 8)...)
				binary.BigEndian.PutUint64(entryBuffer[len(entryBuffer)-8:], uint64(expireTime))

				// Serialize the data
				dataLength := uint32(len(entry.Data))
				entryBuffer = append(entryBuffer, make([]byte, 4)...)
				binary.BigEndian.PutUint32(entryBuffer[len(entryBuffer)-4:], dataLength)
				entryBuffer = append(entryBuffer, entry.Data...)

				if _, err := gzipWriter.Write(entryBuffer); err != nil {
					sh.mutex.RUnlock()
					return err
				}

			}
		}
		sh.mutex.RUnlock()

	}

	// Flush the buffered writer to ensure all data is written to the file
	if err := bufferedWriter.Flush(); err != nil {
		return err
	}

	return nil
}
