package simplegokv

import (
	"bytes"
	"encoding/gob"
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
