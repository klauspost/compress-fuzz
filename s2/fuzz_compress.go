package fuzzs2

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/klauspost/compress/s2"
)

func FuzzCompress(data []byte) int {
	// Test block.
	comp := s2.Encode(nil, data)
	decoded, err := s2.Decode(nil, comp)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, decoded) {
		panic("block decoder mismatch")
	}
	if mel := s2.MaxEncodedLen(len(data)); len(comp) > mel {
		panic(fmt.Errorf("s2.MaxEncodedLen Exceed: input: %d, mel: %d, got %d", len(data), mel, len(comp)))
	}
	// Test writer:
	var buf bytes.Buffer
	enc := s2.NewWriter(&buf, s2.WriterConcurrency(2))
	defer enc.Close()
	n, err := enc.Write(data)
	if err != nil {
		panic(err)
	}
	if n != len(data) {
		panic(fmt.Errorf("Write: Short write, want %d, got %d", len(data), n))
	}
	err = enc.Flush()
	if err != nil {
		panic(err)
	}
	comp = buf.Bytes()
	dec := s2.NewReader(&buf)
	got, err := ioutil.ReadAll(dec)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, got) {
		panic("block (reset) decoder mismatch")
	}

	// Test Reset on both and use ReadFrom instead.
	input := bytes.NewBuffer(data)
	buf = bytes.Buffer{}
	enc.Reset(&buf)
	n2, err := enc.ReadFrom(input)
	if err != nil {
		panic(err)
	}
	if n2 != int64(len(data)) {
		panic(fmt.Errorf("ReadFrom: Short read, want %d, got %d", len(data), n2))
	}
	err = enc.Flush()
	if err != nil {
		panic(err)
	}
	dec.Reset(&buf)
	got, err = ioutil.ReadAll(dec)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, got) {
		panic("frame (reset) decoder mismatch")
	}

	return 1
}
