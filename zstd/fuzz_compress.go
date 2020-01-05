package fuzzzstd

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/klauspost/compress/zstd"
)

const (
	speedNotSet zstd.EncoderLevel = iota

	// SpeedFastest will choose the fastest reasonable compression.
	// This is roughly equivalent to the fastest Zstandard mode.
	SpeedFastest

	// SpeedDefault is the default "pretty fast" compression option.
	// This is roughly equivalent to the default Zstandard mode (level 3).
	SpeedDefault

	// speedLast should be kept as the last actual compression option.
	// The is not for external usage, but is used to keep track of the valid options.
	speedLast

	// SpeedBetterCompression will (in the future) yield better compression than the default,
	// but at approximately 4x the CPU usage of the default.
	// For now this is not implemented.
	SpeedBetterCompression = SpeedDefault

	// SpeedBestCompression will choose the best available compression option.
	// For now this is not implemented.
	SpeedBestCompression = SpeedDefault
)

var dec *zstd.Decoder
var encs [speedLast]*zstd.Encoder
var mu sync.Mutex
var once sync.Once

func initEnc() {
	var err error
	dec, err = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		panic(err)
	}
	for level := zstd.EncoderLevel(speedNotSet + 1); level < speedLast; level++ {
		encs[level], err = zstd.NewWriter(nil, zstd.WithEncoderCRC(false), zstd.WithEncoderLevel(level), zstd.WithEncoderConcurrency(2), zstd.WithWindowSize(128<<10), zstd.WithZeroFrames(true))
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

	for level := zstd.EncoderLevel(speedNotSet + 1); level < speedLast; level++ {
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
