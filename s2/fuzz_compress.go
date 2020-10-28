package fuzzs2

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/klauspost/compress/s2"
	"github.com/klauspost/compress/snappy"
)

var dec *s2.Reader
var enc *s2.Writer
var encBetter *s2.Writer
var once sync.Once

func initEnc() {
	dec = s2.NewReader(nil)
	enc = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBlockSize(128<<10))
	encBetter = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBetterCompression(), s2.WriterBlockSize(512<<10))
}

func FuzzCompress(data []byte) int {
	data = data[:len(data):len(data)]
	once.Do(initEnc)
	// Test block.
	comp := s2.Encode(make([]byte, 0, len(data)/2), data)
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

	comp = s2.EncodeBetter(make([]byte, s2.MaxEncodedLen(len(data))), data)
	decoded, err = s2.Decode(nil, comp)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, decoded) {
		panic("block decoder mismatch")
	}
	if mel := s2.MaxEncodedLen(len(data)); len(comp) > mel {
		panic(fmt.Errorf("MaxEncodedLen Exceed: input: %d, mel: %d, got %d", len(data), mel, len(comp)))
	}

	// Test writer and use "better":
	var buf bytes.Buffer
	encBetter.Reset(&buf)
	n, err := encBetter.Write(data)
	if err != nil {
		panic(err)
	}
	if n != len(data) {
		panic(fmt.Errorf("Write: Short write, want %d, got %d", len(data), n))
	}
	err = encBetter.Close()
	if err != nil {
		panic(err)
	}
	// Calling close twice should not affect anything.
	err = encBetter.Close()
	if err != nil {
		panic(err)
	}
	comp = buf.Bytes()
	if len(comp)%255 != 0 {
		panic(fmt.Errorf("wanted size to be mutiple of %d, got size %d with remainder %d", 255, len(comp), len(comp)%255))
	}
	dec.Reset(&buf)
	got, err := ioutil.ReadAll(dec)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, got) {
		panic("block (reset) decoder mismatch")
	}

	// Test Reset on both and use ReadFrom and EncodeBuffer instead.
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
	err = enc.EncodeBuffer(data)
	if err != nil {
		panic(err)
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
	if !bytes.Equal(data, got[:len(data)]) {
		panic("frame (reset) decoder mismatch, part 1")
	}
	if !bytes.Equal(data, got[len(data):]) {
		panic("frame (reset) decoder mismatch, part 2")
	}

	// Finally test stateless snappy.
	comp = s2.EncodeSnappy(comp, data)
	got, err = snappy.Decode(make([]byte, len(data)), comp)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, got) {
		panic("snappy block decoder mismatch")
	}
	if mel := s2.MaxEncodedLen(len(data)); len(comp) > mel {
		panic(fmt.Errorf("snappy block, s2.MaxEncodedLen Exceed: input: %d, mel: %d, got %d", len(data), mel, len(comp)))
	}

	return 1
}
