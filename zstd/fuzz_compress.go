package fuzzzstd

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/klauspost/compress/zstd"
)

var dec *zstd.Decoder
var encs [zstd.SpeedBestCompression + 1]*zstd.Encoder
var once sync.Once

func initEnc() {
	var err error
	dec, err = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		panic(err)
	}
	for level := zstd.SpeedFastest; level <= zstd.SpeedBestCompression; level++ {
		encs[level], err = zstd.NewWriter(nil, zstd.WithEncoderCRC(true), zstd.WithEncoderLevel(level), zstd.WithEncoderConcurrency(2), zstd.WithWindowSize(128<<10), zstd.WithZeroFrames(true))
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
	}
	return 1
}
