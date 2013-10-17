// Copyright 2013 Frederik Zipp.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
)

type position struct {
	line, col int
}

type selection struct {
	start, end position
}

// parseSelection parses a selection like "startLine.startCol-endLine.endCol".
func parseSelection(str string) (s selection, err error) {
	_, err = fmt.Sscanf(str, "%d.%d-%d.%d",
		&s.start.line, &s.start.col,
		&s.end.line, &s.end.col)
	return s, err
}

// byteOffsetsIn converts a selection to "startByte:endByte".
func (s selection) byteOffsetsIn(b []byte) string {
	return fmt.Sprintf("%d:%d",
		s.start.byteOffsetIn(b),
		s.end.byteOffsetIn(b)+1)
}

func (p position) byteOffsetIn(b []byte) int {
	return nthIndexByte(b, '\n', p.line-1) + p.col
}

func nthIndexByte(s []byte, c byte, n int) int {
	var count, idx int
	for count = 0; count < n; count++ {
		i := bytes.IndexByte(s, c) + 1
		idx += i
		if i == 0 {
			break
		}
		s = s[i:]
	}
	if count == n {
		return idx - 1
	}
	return -1
}
