package fuzzzstd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/klauspost/compress/zstd"
)

var dec *zstd.Decoder
var encs [zstd.SpeedBestCompression + 1]*zstd.Encoder
var encsD [zstd.SpeedBestCompression + 1]*zstd.Encoder
var once sync.Once

func initEnc() {
	dict, err := ioutil.ReadFile("d0.dict")
	if err != nil {
		panic(err)
	}
	dec, err = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1), zstd.WithDecoderDicts(dict))
	if err != nil {
		panic(err)
	}
	for level := zstd.SpeedFastest; level <= zstd.SpeedBestCompression; level++ {
		encs[level], err = zstd.NewWriter(nil, zstd.WithEncoderCRC(true), zstd.WithEncoderLevel(level), zstd.WithEncoderConcurrency(2), zstd.WithWindowSize(128<<10), zstd.WithZeroFrames(true))
		encsD[level], err = zstd.NewWriter(nil, zstd.WithEncoderCRC(true), zstd.WithEncoderLevel(level), zstd.WithEncoderConcurrency(2), zstd.WithWindowSize(128<<10), zstd.WithZeroFrames(true), zstd.WithEncoderDict(dict))
	}
}

func FuzzCompress(data []byte) int {
	once.Do(initEnc)
	// Run test against out decoder
	var dst bytes.Buffer

	// Create a buffer that will usually be too small.
	var bufSize = len(data)
	if bufSize > 2 {
		// Make deterministic size
		bufSize = int(data[0]) | int(data[1])<<8
		if bufSize >= len(data) {
			bufSize = len(data) / 2
		}
	}

	for level := zstd.SpeedFastest; level <= zstd.SpeedBestCompression; level++ {
		enc := encs[level]
		dst.Reset()
		enc.Reset(&dst)
		n, err := enc.Write(data)
		if err != nil {
			panic(err)
		}
		if n != len(data) {
			panic(fmt.Sprintln("Level", level, "Short write, got:", n, "want:", len(data)))
		}

		encoded := enc.EncodeAll(data, make([]byte, 0, bufSize))
		got, err := dec.DecodeAll(encoded, make([]byte, 0, bufSize))
		if err != nil {
			panic(fmt.Sprintln("Level", level, "DecodeAll error:", err, "\norg:", len(data), "\nencoded", len(encoded)))
		}
		if !bytes.Equal(got, data) {
			panic(fmt.Sprintln("Level", level, "DecodeAll output mismatch\n", len(got), "org: \n", len(data), "(want)", "\nencoded:", len(encoded)))
		}

		err = enc.Close()
		if err != nil {
			panic(fmt.Sprintln("Level", level, "Close (buffer) error:", err))
		}
		encoded2 := dst.Bytes()
		if !bytes.Equal(encoded, encoded2) {
			got, err = dec.DecodeAll(encoded2, got[:0])
			if err != nil {
				panic(fmt.Sprintln("Level", level, "DecodeAll (buffer) error:", err, "\norg:", len(data), "\nencoded", len(encoded2)))
			}
			if !bytes.Equal(got, data) {
				panic(fmt.Sprintln("Level", level, "DecodeAll (buffer) output mismatch\n", len(got), "org: \n", len(data), "(want)", "\nencoded:", len(encoded2)))
			}
		}

		enc = encsD[level]
		dst.Reset()
		enc.Reset(&dst)
		n, err = enc.Write(data)
		if err != nil {
			panic(err)
		}
		if n != len(data) {
			panic(fmt.Sprintln("Dict Level", level, "Short write, got:", n, "want:", len(data)))
		}

		encoded = enc.EncodeAll(data, encoded[:0])
		got, err = dec.DecodeAll(encoded, got[:0])
		if err != nil {
			panic(fmt.Sprintln("Dict Level", level, "DecodeAll error:", err, "\norg:", len(data), "\nencoded", len(encoded)))
		}
		if !bytes.Equal(got, data) {
			panic(fmt.Sprintln("Dict Level", level, "DecodeAll output mismatch\n", len(got), "org: \n", len(data), "(want)", "\nencoded:", len(encoded)))
		}

		err = enc.Close()
		if err != nil {
			panic(fmt.Sprintln("Dict Level", level, "Close (buffer) error:", err))
		}
		encoded2 = dst.Bytes()
		if !bytes.Equal(encoded, encoded2) {
			got, err = dec.DecodeAll(encoded2, got[:0])
			if err != nil {
				panic(fmt.Sprintln("Dict Level", level, "DecodeAll (buffer) error:", err, "\norg:", len(data), "\nencoded", len(encoded2)))
			}
			if !bytes.Equal(got, data) {
				panic(fmt.Sprintln("Dict Level", level, "DecodeAll (buffer) output mismatch\n", len(got), "org: \n", len(data), "(want)", "\nencoded:", len(encoded2)))
			}
		}
	}
	return 1
}

// FuzzCompressSimple will test a single level and only encoding all.
// Can be good for generating corpus or checking a specific compression level.
func FuzzCompressSimple(data []byte) int {
	once.Do(initEnc)

	// Create a buffer that will usually be too small.
	var bufSize = len(data)
	if bufSize > 2 {
		// Make deterministic size
		bufSize = int(data[0]) | int(data[1])<<8
		if bufSize >= len(data) {
			bufSize = len(data) / 2
		}
	}
	const level = zstd.SpeedBestCompression
	enc := encs[level]

	encoded := enc.EncodeAll(data, make([]byte, 0, bufSize))
	got, err := dec.DecodeAll(encoded, make([]byte, 0, len(data)))
	if err != nil {
		panic(fmt.Sprintln("Level", level, "DecodeAll error:", err, "\norg:", len(data), "\nencoded", len(encoded)))
	}
	if !bytes.Equal(got, data) {
		panic(fmt.Sprintln("Level", level, "DecodeAll output mismatch\n", len(got), "org: \n", len(data), "(want)", "\nencoded:", len(encoded)))
	}

	enc = encsD[level]

	encoded = enc.EncodeAll(data, encoded[:0])
	got, err = dec.DecodeAll(encoded, got[:0])
	if err != nil {
		panic(fmt.Sprintln("Dict Level", level, "DecodeAll error:", err, "\norg:", len(data), "\nencoded", len(encoded)))
	}
	if !bytes.Equal(got, data) {
		panic(fmt.Sprintln("Dict Level", level, "DecodeAll output mismatch\n", len(got), "org: \n", len(data), "(want)", "\nencoded:", len(encoded)))
	}

	return 1
}
