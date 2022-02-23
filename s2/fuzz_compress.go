package fuzzs2

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/golang/snappy"
	"github.com/klauspost/compress/s2"
)

var dec *s2.Reader
var enc *s2.Writer
var encBetter *s2.Writer
var encBest *s2.Writer
var once sync.Once

func initEnc() {
	dec = s2.NewReader(nil)
	enc = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBlockSize(16<<10))
	encBetter = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBetterCompression(), s2.WriterBlockSize(64<<10))
	encBest = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBetterCompression(), s2.WriterBlockSize(512<<10), s2.WriterFlushOnWrite())
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

	comp = s2.EncodeBetter(comp, data)
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

	comp = s2.EncodeBest(comp, data)
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

	// Test writer and use "best":
	encBest.Reset(&buf)
	n64, err := io.CopyBuffer(io.MultiWriter(encBest), bytes.NewBuffer(data), make([]byte, 4000))
	n = int(n64)
	if err != nil {
		panic(err)
	}
	if n != len(data) {
		panic(fmt.Errorf("Write: Short write, want %d, got %d", len(data), n))
	}
	err = encBest.Close()
	if err != nil {
		panic(err)
	}
	// Calling close twice should not affect anything.
	err = encBest.Close()
	if err != nil {
		panic(err)
	}
	comp = buf.Bytes()
	if len(comp)%255 != 0 {
		panic(fmt.Errorf("wanted size to be mutiple of %d, got size %d with remainder %d", 255, len(comp), len(comp)%255))
	}
	dec.Reset(&buf)
	got, err = ioutil.ReadAll(dec)
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

	return 1
}

var decS *snappy.Reader

func initEncSnappy() {
	decS = snappy.NewReader(nil)
	enc = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBlockSize(16<<10), s2.WriterSnappyCompat())
	encBetter = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBetterCompression(), s2.WriterSnappyCompat())
	encBest = s2.NewWriter(nil, s2.WriterConcurrency(2), s2.WriterPadding(255), s2.WriterBetterCompression(), s2.WriterSnappyCompat())
}

func FuzzCompressSnappy(data []byte) int {
	data = data[:len(data):len(data)]
	once.Do(initEncSnappy)
	// Test block.
	comp := s2.EncodeSnappy(make([]byte, 0, len(data)/2), data)
	decoded, err := snappy.Decode(nil, comp)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, decoded) {
		panic("block decoder mismatch")
	}
	if mel := s2.MaxEncodedLen(len(data)); len(comp) > mel {
		panic(fmt.Errorf("s2.MaxEncodedLen Exceed: input: %d, mel: %d, got %d", len(data), mel, len(comp)))
	}

	comp = s2.EncodeSnappyBetter(comp, data)
	decoded, err = snappy.Decode(nil, comp)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, decoded) {
		panic("block decoder mismatch")
	}
	if mel := s2.MaxEncodedLen(len(data)); len(comp) > mel {
		panic(fmt.Errorf("MaxEncodedLen Exceed: input: %d, mel: %d, got %d", len(data), mel, len(comp)))
	}

	comp = s2.EncodeSnappyBest(comp, data)
	decoded, err = snappy.Decode(nil, comp)
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
	decS.Reset(&buf)
	got, err := ioutil.ReadAll(decS)
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
	decS.Reset(&buf)
	got, err = ioutil.ReadAll(decS)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, got[:len(data)]) {
		panic("frame (reset) decoder mismatch, part 1")
	}
	if !bytes.Equal(data, got[len(data):]) {
		panic("frame (reset) decoder mismatch, part 2")
	}

	return 1
}

func FuzzCompressSingle(data []byte) int {
	maxLen := s2.MaxEncodedLen(len(data))
	dst := make([]byte, maxLen+1, maxLen+1)
	// Check if we write beyond MaxEncodedLen.
	dst[maxLen] = 0x39
	// switch this to another
	comp := s2.EncodeBetter(dst[:], data)
	if len(comp) > maxLen {
		panic("too large output")
	}
	for _, v := range dst[maxLen : maxLen+1] {
		if v != 0x39 {
			panic(fmt.Sprintf("wrote extra out-of-bounds, want 0x39, got %x, %d -> %v (max %d)", v, len(data), len(comp), maxLen))
		}
	}
	decoded, err := s2.Decode(nil, comp)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, decoded) {
		panic("block decoder mismatch")
	}
	if mel := s2.MaxEncodedLen(len(data)); len(comp) > mel {
		panic(fmt.Errorf("MaxEncodedLen Exceed: input: %d, mel: %d, got %d", len(data), mel, len(comp)))
	}
	return 1
}
