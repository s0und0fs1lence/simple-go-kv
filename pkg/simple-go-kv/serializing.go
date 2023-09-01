package simplegokv

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"io/ioutil"
)

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

func (k kvStore) compress(in []byte) ([]byte, error) {
	var compressedData bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedData)
	defer gzipWriter.Close()

	_, err := gzipWriter.Write(in)
	if err != nil {
		return nil, err
	}
	return compressedData.Bytes(), nil
}

func (k kvStore) decompress(in []byte) ([]byte, error) {
	compressedDataReader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	defer compressedDataReader.Close()

	decompressedData, err := ioutil.ReadAll(compressedDataReader)
	if err != nil {
		return nil, err
	}

	return decompressedData, nil
}
