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
	// Test writer and use "better":
	var buf bytes.Buffer
	enc := s2.NewWriter(&buf, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBetterCompression())
	defer enc.Close()
	n, err := enc.Write(data)
	if err != nil {
		panic(err)
	}
	if n != len(data) {
		panic(fmt.Errorf("Write: Short write, want %d, got %d", len(data), n))
	}
	err = enc.Close()
	if err != nil {
		panic(err)
	}
	// Calling close twice should not affect anything.
	err = enc.Close()
	if err != nil {
		panic(err)
	}
	comp = buf.Bytes()
	if len(comp)%255 != 0 {
		panic(fmt.Errorf("wanted size to be mutiple of %d, got size %d with remainder %d", 255, len(comp), len(comp)%255))
	}
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
	err = enc.Close()
	if err != nil {
		panic(err)
	}
	if buf.Len()%255 != 0 {
		panic(fmt.Errorf("wanted size to be mutiple of %d, got size %d with remainder %d", 255, buf.Len(), buf.Len()%255))
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
