// Copyright 2015 go-fuzz project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package fuzzflate

import (
	"bytes"
	"io/ioutil"
	"strconv"

	"github.com/klauspost/compress/flate"
)

// Fuzz tests all encoding levels
func Fuzz(data []byte) int {
	// Skip large blocks.
	if len(data) > 1<<20 {
		return 0
	}
	reader := flate.NewReader(nil)
	readerReset := reader.(flate.Resetter)
	{
		level := "stateless"
		msg := "level " + level + ":"
		buf := new(bytes.Buffer)
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
		data2, err = ioutil.ReadAll(reader)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data, data2) {
			panic(msg + "not equal")
		}
	}
	for level := -2; level <= 9; level++ {
		msg := "level " + strconv.Itoa(level) + ":"
		buf := new(bytes.Buffer)
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
		// Do it again...
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
		data2, err = ioutil.ReadAll(reader)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data, data2) {
			panic(msg + "not equal")
		}
	}
	return 1
}
