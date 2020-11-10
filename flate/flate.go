// Copyright 2015 go-fuzz project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package fuzzflate

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/klauspost/compress/flate"
)

// Fuzz tests all encoding levels
func Fuzz(data []byte) int {
	const decodeOnly = false
	// Skip large blocks.
	if len(data) > 1<<20 || (decodeOnly && len(data) > 32<<10) {
		return 0
	}

	reader := flate.NewReader(nil)
	decBuf := bytes.NewBuffer(make([]byte, 0, len(data)))
	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	readerReset := reader.(flate.Resetter)
	if !decodeOnly {
		level := "stateless"
		msg := "level " + level + ":"
		buf.Reset()
		fw := flate.NewStatelessWriter(buf)
		n, err := fw.Write(data)
		if n != len(data) {
			panic(msg + "short write")
		}
		if err != nil {
			panic(msg + err.Error())
		}
		err = fw.Close()
		if err != nil {
			panic(msg + err.Error())
		}
		err = readerReset.Reset(buf, nil)
		if err != nil {
			panic(msg + err.Error())
		}
		data2, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data, data2) {
			panic(msg + "not equal")
		}
		// Do it again...
		msg = "level " + level + " (reset):"
		buf.Reset()
		fw = flate.NewStatelessWriter(buf)
		n, err = fw.Write(data)
		if n != len(data) {
			panic(msg + "short write")
		}
		if err != nil {
			panic(msg + err.Error())
		}
		err = fw.Close()
		if err != nil {
			panic(msg + err.Error())
		}
		err = readerReset.Reset(buf, nil)
		if err != nil {
			panic(msg + err.Error())
		}
		decBuf.Reset()
		_, err = reader.(io.WriterTo).WriteTo(decBuf)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data, decBuf.Bytes()) {
			panic(msg + "not equal")
		}
	}

	for level := -2; level <= 9; level++ {
		if decodeOnly {
			break
		}
		msg := "level " + strconv.Itoa(level) + ":"
		buf.Reset()
		fw, err := flate.NewWriter(buf, level)
		if err != nil {
			panic(msg + err.Error())
		}
		n, err := fw.Write(data)
		if n != len(data) {
			panic(msg + "short write")
		}
		if err != nil {
			panic(msg + err.Error())
		}
		err = fw.Close()
		if err != nil {
			panic(msg + err.Error())
		}
		err = readerReset.Reset(buf, nil)
		if err != nil {
			panic(msg + err.Error())
		}
		data2, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data, data2) {
			panic(msg + "not equal")
		}
		// Do it again... Use WriteTo on inflate.
		msg = "level " + strconv.Itoa(level) + " (reset):"
		buf.Reset()
		fw.Reset(buf)
		n, err = fw.Write(data)
		if n != len(data) {
			panic(msg + "short write")
		}
		if err != nil {
			panic(msg + err.Error())
		}
		err = fw.Close()
		if err != nil {
			panic(msg + err.Error())
		}
		err = readerReset.Reset(buf, nil)
		if err != nil {
			panic(msg + err.Error())
		}
		decBuf.Reset()
		_, err = reader.(io.WriterTo).WriteTo(decBuf)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data, decBuf.Bytes()) {
			panic(msg + "not equal")
		}
	}

	// Try decode raw data, should just not crash.
	decBuf.Reset()
	err := readerReset.Reset(bytes.NewBuffer(data), data)
	if err != nil {
		reader.(io.WriterTo).WriteTo(ioutil.Discard)
	} else if decodeOnly {
		return 0
	}
	return 1
}
